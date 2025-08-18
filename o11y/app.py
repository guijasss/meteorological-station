import docker
import time
from tabulate import tabulate

# Inicializa cliente Docker
client = docker.from_env()

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
        cpu_percent = (cpu_delta / system_delta) * num_cpus * 100.0 if system_delta > 0 else 0.0

        # Memória
        mem_used = stats["memory_stats"]["usage"] / (1024 * 1024)
        mem_limit = stats["memory_stats"]["limit"] / (1024 * 1024)
        mem_percent = (mem_used / mem_limit * 100) if mem_limit > 0 else 0.0

        # Rede
        networks = stats.get("networks", {})
        rx = sum(n["rx_bytes"] for n in networks.values()) / (1024*1024)
        tx = sum(n["tx_bytes"] for n in networks.values()) / (1024*1024)

        # I/O de bloco
        blkio = stats.get("blkio_stats", {}).get("io_service_bytes_recursive") or []
        io_read = sum(x["value"] for x in blkio if x.get("op") == "Read") / (1024*1024)
        io_write = sum(x["value"] for x in blkio if x.get("op") == "Write") / (1024*1024)

        # Reinícios
        restarts = c.attrs.get("RestartCount", 0)

        metrics_list.append({
            "Container": c.name,
            "CPU %": f"{cpu_percent:.2f}%",
            "MEM USAGE / LIMIT": f"{mem_used:.2f}MiB / {mem_limit:.0f}MiB",
            "MEM %": f"{mem_percent:.2f}%",
            "NET I/O": f"{rx:.2f}MB / {tx:.2f}MB",
            "BLOCK I/O": f"{io_read:.2f}MB / {io_write:.2f}MB",
            "Restarts": restarts
        })

    return metrics_list

# Função para exibir tabela no terminal
def show_metrics():
    while True:
        metrics = get_container_metrics()
        print("\033c", end="")  # limpa terminal
        print(f"📊 Métricas dos Containers - Atualizado em {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print(tabulate(metrics, headers="keys", tablefmt="grid"))
        time.sleep(5)

if __name__ == "__main__":
    show_metrics()
