#!/bin/bash

# Verifica se o nome do container foi passado
if [ -z "$1" ]; then
  echo "Uso: $0 <nome_do_container>"
  exit 1
fi

CONTAINER_NAME=$1

# Verifica se o container existe
if ! docker inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
  echo "Erro: container '$CONTAINER_NAME' não encontrado."
  exit 1
fi

# Pega todos os IPs das redes conectadas ao container
IPS=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}' "$CONTAINER_NAME")

if [ -z "$IPS" ]; then
  echo "Nenhum IP encontrado para o container '$CONTAINER_NAME'."
  exit 1
fi

echo "IP(s) do container '$CONTAINER_NAME': $IPS"
