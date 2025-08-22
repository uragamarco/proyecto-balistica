# Integración Python-Go en Proyecto Balística

## Introducción

Este documento describe la integración entre Python y Go implementada en el Proyecto Balística. Esta integración permite aprovechar las capacidades de procesamiento de imágenes de Python junto con la eficiencia y concurrencia de Go.

## Arquitectura

La integración se basa en un enfoque de comunicación entre procesos (IPC) donde:

1. Go actúa como el lenguaje principal para la aplicación
2. Python se utiliza para tareas específicas de procesamiento de imágenes y extracción de características
3. La comunicación se realiza mediante la ejecución de scripts Python desde Go y el intercambio de datos a través de JSON

## Componentes Principales

### PythonService (Go)

Definido en `pkg/integration/PythonService.go`, este servicio gestiona la comunicación con los scripts Python:

- Inicializa el servicio Python
- Proporciona métodos para extraer características de imágenes
- Implementa verificaciones de salud (health checks)
- Maneja errores y timeouts

### RPC Bridge (Go)

Definido en `pkg/integration/rpc_bridge.go`, este componente:

- Ejecuta comandos Python
- Parsea las respuestas JSON
- Maneja errores de comunicación

### Feature Extractor (Python)

Definido en `py_services/feature_extractor.py`, este script:

- Proporciona una API Flask para extracción de características
- Implementa funciones para calcular momentos de Hu y otras características
- Puede ejecutarse como servicio web o como script de línea de comandos

## Flujo de Comunicación

1. La aplicación Go recibe una solicitud para procesar una imagen
2. Go preprocesa la imagen y la guarda en un archivo temporal
3. Go llama al servicio Python para extraer características avanzadas
4. Python procesa la imagen y extrae características (momentos de Hu, área de contorno, etc.)
5. Python devuelve los resultados en formato JSON
6. Go parsea la respuesta JSON y la integra con sus propios resultados

## Manejo de Errores

La integración incluye un manejo de errores robusto:

- Verificación de existencia de archivos antes de procesarlos
- Timeouts para evitar bloqueos indefinidos
- Parsing de respuestas de error desde Python
- Logging detallado para facilitar la depuración

## Ejemplo de Uso

### Extracción de Características

```go
// En Go
features, err := pythonService.ExtractFeatures(imagePath)
if err != nil {
    // Manejar error
}

// Usar las características extraídas
huMoment1 := features["hu_moment_1"]
contourArea := features["contour_area"]
```

```python
# En Python (feature_extractor.py)
def calculate_hu_moments(image_data):
    # Procesar imagen y calcular momentos de Hu
    # ...
    return {
        "hu_moment_1": hu[0],
        "hu_moment_2": hu[1],
        # ...
        "contour_area": area,
        "contour_length": length,
        "striation_density": striations
    }
```

## Modo Dual de Feature Extractor

El script `feature_extractor.py` puede funcionar en dos modos:

1. **Modo Servidor Flask**: Para desarrollo y pruebas interactivas
   ```bash
   python feature_extractor.py
   ```

2. **Modo Línea de Comandos**: Para integración con Go
   ```bash
   python feature_extractor.py --image /ruta/a/imagen.png
   ```

## Verificación de Salud (Health Check)

La integración incluye un mecanismo de verificación de salud:

```go
// En Go
status, err := pythonService.HealthCheck()
if err != nil {
    log.Error("Servicio Python no disponible", zap.Error(err))
}
```

## Mejores Prácticas

1. **Manejo de Recursos**: Liberar recursos adecuadamente (archivos temporales, etc.)
2. **Timeouts**: Implementar timeouts para todas las operaciones de IPC
3. **Validación de Entrada/Salida**: Validar datos antes y después de la comunicación
4. **Logging**: Mantener logs detallados para facilitar la depuración
5. **Manejo de Errores**: Implementar manejo de errores robusto en ambos lados

## Conclusión

La integración Python-Go proporciona una solución flexible y potente para combinar las fortalezas de ambos lenguajes en el procesamiento de imágenes balísticas.