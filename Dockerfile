# Build stage - Go 1.23.0 con Alpine
FROM golang:1.23.0-alpine3.19 AS builder

WORKDIR /app
COPY . .

# Instalar dependencias necesarias para OpenCV
RUN apk update && apk add --no-cache \
    build-base \
    cmake \
    pkgconf \
    opencv-dev

# Descargar dependencias y construir
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /proyecto-balistica cmd/main.go

# Runtime stage - Alpine ligero
FROM alpine:3.19

# Instalar solo dependencias de ejecución
RUN apk update && apk add --no-cache \
    libstdc++ \
    libgcc \
    opencv

WORKDIR /app

# Copiar artefactos de construcción
COPY --from=builder /proyecto-balistica .
COPY configs ./configs
COPY .env .

# Configurar permisos
RUN chmod +x proyecto-balistica

EXPOSE 8080

CMD ["./proyecto-balistica"]