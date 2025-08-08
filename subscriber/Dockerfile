FROM golang:1.23.0-alpine

WORKDIR /app

# Copia go.mod e go.sum do contexto raiz para /app
COPY ../go.mod ./

RUN go mod download

COPY . .

RUN go build -o subscriber
