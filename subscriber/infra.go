package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type QuestDBTCPClient struct {
	address  string
	connPool chan net.Conn
	poolSize int
	mu       sync.RWMutex
	metrics  ClientMetrics
}

type ClientMetrics struct {
	Sent    int64
	Failed  int64
	Pending int64
}

func NewQuestDBTCPClient(address string, poolSize int) *QuestDBTCPClient {
	client := &QuestDBTCPClient{
		address:  address,
		connPool: make(chan net.Conn, poolSize),
		poolSize: poolSize,
	}

	for i := 0; i < poolSize; i++ {
		if conn, err := net.DialTimeout("tcp", address, 5*time.Second); err == nil {
			client.connPool <- conn
		}
	}

	fmt.Printf("✅ Pool TCP criado com %d conexões para %s\n", len(client.connPool), address)
	return client
}

func (c *QuestDBTCPClient) getConnection() net.Conn {
	select {
	case conn := <-c.connPool:
		return conn
	default:
		if conn, err := net.DialTimeout("tcp", c.address, 5*time.Second); err == nil {
			return conn
		}
		return nil
	}
}

func (c *QuestDBTCPClient) returnConnection(conn net.Conn) {
	select {
	case c.connPool <- conn:
	default:
		conn.Close()
	}
}

func (c *QuestDBTCPClient) SendBatch(events []SensorEvent) error {
	if len(events) == 0 {
		return nil
	}

	conn := c.getConnection()
	if conn == nil {
		c.mu.Lock()
		c.metrics.Failed += int64(len(events))
		c.mu.Unlock()
		return fmt.Errorf("não foi possível obter conexão")
	}

	var batch strings.Builder
	for _, event := range events {
		line := fmt.Sprintf(
			"readings,station=%s,sensor=%s value=%f %d\n",
			sanitizeTag(event.Station),
			sanitizeTag(event.Sensor),
			event.Value,
			event.Timestamp*1_000_000_000,
		)
		batch.WriteString(line)
	}

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	_, err := conn.Write([]byte(batch.String()))
	if err != nil {
		conn.Close()
		c.mu.Lock()
		c.metrics.Failed += int64(len(events))
		c.mu.Unlock()
		return fmt.Errorf("erro ao enviar batch: %v", err)
	}

	c.returnConnection(conn)
	c.mu.Lock()
	c.metrics.Sent += int64(len(events))
	c.mu.Unlock()

	return nil
}

func (c *QuestDBTCPClient) GetMetrics() ClientMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics
}

func (c *QuestDBTCPClient) Close() {
	close(c.connPool)
	for conn := range c.connPool {
		conn.Close()
	}
}

// Processador assíncrono
type AsyncProcessor struct {
	tcpClient     *QuestDBTCPClient
	buffer        chan SensorEvent
	batchSize     int
	flushInterval time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewAsyncProcessor(bufferSize, batchSize int) *AsyncProcessor {
	processor := &AsyncProcessor{
		tcpClient:     NewQuestDBTCPClient("questdb:9009", 5),
		buffer:        make(chan SensorEvent, bufferSize),
		batchSize:     batchSize,
		flushInterval: 100 * time.Millisecond,
		stopCh:        make(chan struct{}),
	}

	processor.wg.Add(1)
	go processor.process()

	return processor
}

func (p *AsyncProcessor) SendEvent(event SensorEvent) {
	select {
	case p.buffer <- event:
	default:
		fmt.Println("⚠️ Buffer QuestDB cheio, descartando evento")
	}
}

func (p *AsyncProcessor) process() {
	defer p.wg.Done()
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	var batch []SensorEvent

	for {
		select {
		case event := <-p.buffer:
			batch = append(batch, event)

			if len(batch) >= p.batchSize {
				p.flushBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				p.flushBatch(batch)
				batch = nil
			}

		case <-p.stopCh:
			// Enviar batch final antes de parar
			if len(batch) > 0 {
				p.flushBatch(batch)
			}
			return
		}
	}
}

func (p *AsyncProcessor) flushBatch(batch []SensorEvent) {
	if err := p.tcpClient.SendBatch(batch); err != nil {
		fmt.Printf("❌ Erro ao enviar batch para QuestDB: %v\n", err)
	}
}

func (p *AsyncProcessor) GetMetrics() (ClientMetrics, int) {
	metrics := p.tcpClient.GetMetrics()
	metrics.Pending = int64(len(p.buffer))
	return metrics, len(p.buffer)
}

func (p *AsyncProcessor) Stop() {
	close(p.stopCh)
	p.wg.Wait()
	p.tcpClient.Close()
}

func sanitizeTag(tag string) string {
	tag = strings.ReplaceAll(tag, " ", "_")
	tag = strings.ReplaceAll(tag, ",", "_")
	tag = strings.ReplaceAll(tag, "=", "_")
	return tag
}
