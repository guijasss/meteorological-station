import docker
import time
from tabulate import tabulate

# Inicializa cliente Docker
client = docker.from_env()

DEFAULT_PRECISION = 2

# Função para coletar métricas de todos os containers
def get_container_metrics():
    containers = client.containers.list()
    metrics_list = []

    for c in containers:
        stats = c.stats(stream=False)

        # CPU %
        cpu_delta = stats["cpu_stats"]["cpu_usage"]["total_usage"] - stats["precpu_stats"]["cpu_usage"]["total_usage"]
        system_delta = stats["cpu_stats"].get("system_cpu_usage", 0) - stats["precpu_stats"].get("system_cpu_usage", 0)
        percpu = stats["cpu_stats"]["cpu_usage"].get("percpu_usage", [])
        num_cpus = len(percpu) if percpu else 1
        cpu_percent = round((cpu_delta / system_delta) * num_cpus * 100.0, DEFAULT_PRECISION) if system_delta > 0 else 0.0

        # Memória
        mem_used = round(stats["memory_stats"]["usage"] / (1024 * 1024), DEFAULT_PRECISION)
        mem_limit = round(stats["memory_stats"]["limit"] / (1024 * 1024), DEFAULT_PRECISION)
        mem_percent = round((mem_used / mem_limit * 100), DEFAULT_PRECISION) if mem_limit > 0 else 0.0

        # Rede
        networks = stats.get("networks", {})
        rx = round(sum(n["rx_bytes"] for n in networks.values()) / (1024*1024), DEFAULT_PRECISION)
        tx = round(sum(n["tx_bytes"] for n in networks.values()) / (1024*1024), DEFAULT_PRECISION)

        # I/O de bloco
        blkio = stats.get("blkio_stats", {}).get("io_service_bytes_recursive") or []
        io_read = sum(x["value"] for x in blkio if x.get("op") == "Read") / (1024*1024)
        io_write = sum(x["value"] for x in blkio if x.get("op") == "Write") / (1024*1024)

        # Reinícios
        restarts = c.attrs.get("RestartCount", 0)

        metrics_list.append({
            "container": c.name,
            "cpu_percent": cpu_percent,
            "mem_usage": mem_used,
            "mem_limit": mem_limit,
            "mem_percent": mem_percent,
            "rx_bytes": rx,
            "tx_bytes": tx,
            "io_read": io_read,
            "io_write": io_write,
            "restarts": restarts
        })

    return metrics_list

# Função para exibir tabela no terminal
def show_metrics():
    while True:
        metrics = [
            {
                "Container": m.get("container"),
                "CPU %": f'{m.get("cpu_percent")}%',
                "MEM USAGE / LIMIT": f'{m.get("mem_usage")} MiB / {m.get("mem_limit"):.0f}MiB',
                "MEM %": f"{m.get('mem_percent')}%",
                "NET I/O": f"{m.get('rx_bytes')}MB / {m.get('tx_bytes')}MB",
                "BLOCK I/O": f"{m.get('io_read')}MB / {m.get('io_write')}MB",
                "Restarts": m.get("restarts")
            } for m in
            get_container_metrics()
        ]

        print("\033c", end="")  # limpa terminal
        print(f"📊 Métricas dos Containers - Atualizado em {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print(tabulate(metrics, headers="keys", tablefmt="grid"))
        time.sleep(5)

if __name__ == "__main__":
    show_metrics()
