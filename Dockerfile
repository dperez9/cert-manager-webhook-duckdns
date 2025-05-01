# Etapa 1: Construcción
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilación estática
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook -ldflags="-s -w"

# Etapa 2: Imagen final minimalista
FROM gcr.io/distroless/static

COPY --from=builder /app/webhook /webhook

ENTRYPOINT ["/webhook"]
