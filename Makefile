.PHONY: run build clean test lint setup

APP_NAME = proyecto-balistica
BIN_DIR = bin

setup:
	@echo "Configurando el entorno de desarrollo..."
	./scripts/setup.sh

run:
	@echo "Iniciando servidor en modo desarrollo..."
	go run cmd/main.go

build:
	@echo "Compilando aplicación..."
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) cmd/main.go

test:
	@echo "Ejecutando tests..."
	go test -v ./...

lint:
	@echo "Ejecutando linter..."
	golangci-lint run

clean:
	@echo "Limpiando binarios..."
	rm -rf $(BIN_DIR)

docker-build:
	@echo "Construyendo imagen Docker..."
	docker build -t $(APP_NAME) .

docker-run:
	@echo "Ejecutando contenedor Docker..."
	docker run -p 8080:8080 $(APP_NAME)