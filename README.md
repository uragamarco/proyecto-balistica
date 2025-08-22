# Proyecto Balística

## Descripción

Proyecto Balística es una aplicación para el análisis de imágenes balísticas que permite procesar, extraer características y comparar muestras de proyectiles. La aplicación utiliza técnicas de procesamiento de imágenes y aprendizaje automático para identificar patrones en proyectiles balísticos.

## Características

- Procesamiento de imágenes balísticas
- Extracción de características (momentos de Hu, área de contorno, longitud de contorno, estriaciones)
- Análisis de croma y colores dominantes
- Comparación de muestras balísticas
- Integración con servicios Python para extracción avanzada de características

## Requisitos

- Go 1.18 o superior
- Python 3.8 o superior
- OpenCV (para las funcionalidades Python)
- NumPy
- Flask

## Instalación

### Requisitos previos

1. Instalar Go: [https://golang.org/doc/install](https://golang.org/doc/install)
2. Instalar Python: [https://www.python.org/downloads/](https://www.python.org/downloads/)
3. Instalar dependencias de Python:

```bash
pip install -r py_services/requirements.txt
```

### Compilación

```bash
go build -o main ./cmd/main.go
```

## Uso

### Iniciar la aplicación

```bash
./main
```

La aplicación estará disponible en `http://localhost:8080`.

### API Endpoints

- `POST /api/process`: Procesa una imagen balística y extrae características
- `POST /api/compare`: Compara dos muestras balísticas

## Estructura del proyecto

```
├── cmd/                # Punto de entrada de la aplicación
├── configs/            # Archivos de configuración
├── internal/           # Código interno de la aplicación
│   ├── api/            # Manejadores de API
│   ├── app/            # Inicialización de la aplicación
│   ├── config/         # Configuración
│   ├── models/         # Modelos de datos
│   └── services/       # Servicios de la aplicación
├── pkg/                # Paquetes reutilizables
│   └── integration/    # Integración con servicios externos
├── py_services/        # Servicios Python
└── web/                # Interfaz web
```

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - ver el archivo LICENSE para más detalles.