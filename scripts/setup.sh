#!/bin/bash

# Instalar dependencias del sistema
echo "Instalando dependencias del sistema..."
sudo apt install -y build-essential pkg-config libopencv-dev

# Configurar entorno Go
echo "Configurando entorno Go..."
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Instalar herramientas de desarrollo
echo "Instalando herramientas de desarrollo..."
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Descargar dependencias del proyecto
echo "Descargando dependencias..."
go mod download

echo "Configuración completada. Ejecuta 'make run' para iniciar el servidor."