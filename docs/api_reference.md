# API Reference - Proyecto Balística

## Introducción

Este documento describe los endpoints de la API REST proporcionados por el Proyecto Balística. La API permite procesar imágenes balísticas y comparar muestras para análisis forense.

## Base URL

```
http://localhost:8080/api
```

## Endpoints

### Procesar Imagen

Procesa una imagen balística y extrae sus características.

**URL**: `/process`

**Método**: `POST`

**Content-Type**: `multipart/form-data`

**Parámetros**:

| Nombre | Tipo | Descripción |
|--------|------|-------------|
| image  | File | Archivo de imagen a procesar (PNG, JPEG) |

**Respuesta Exitosa**:

- **Código**: 200 OK
- **Content-Type**: `application/json`
- **Cuerpo**:

```json
{
  "features": {
    "hu_moment_1": 0.123456,
    "hu_moment_2": 0.234567,
    "hu_moment_3": 0.345678,
    "hu_moment_4": 0.456789,
    "hu_moment_5": 0.567890,
    "hu_moment_6": 0.678901,
    "hu_moment_7": 0.789012,
    "contour_area": 12345.67,
    "contour_length": 456.78,
    "striation_density": 0.89
  },
  "chroma_data": {
    "dominant_colors": [
      {
        "color": {
          "r": 120,
          "g": 120,
          "b": 120
        },
        "percentage": 0.65
      },
      {
        "color": {
          "r": 200,
          "g": 200,
          "b": 200
        },
        "percentage": 0.35
      }
    ],
    "color_variance": 0.123
  },
  "metadata": {
    "timestamp": "2023-07-15T14:22:10Z",
    "image_hash": "a1b2c3d4e5f6g7h8i9j0",
    "processor_version": "1.4.0",
    "python_features_used": true,
    "confidence": 0.92,
    "filename": "muestra_bala_001.jpg",
    "content_type": "image/jpeg",
    "file_size": 1024567
  }
}
```

**Respuestas de Error**:

- **Código**: 400 Bad Request
  - Cuando la imagen no se proporciona o no es válida
- **Código**: 415 Unsupported Media Type
  - Cuando el formato de imagen no es compatible
- **Código**: 500 Internal Server Error
  - Cuando ocurre un error durante el procesamiento

**Ejemplo de Error**:

```json
{
  "code": 400,
  "message": "Imagen no proporcionada",
  "details": "No se encontró ningún archivo en la solicitud"
}
```

### Comparar Muestras

Compara dos conjuntos de características balísticas para determinar su similitud.

**URL**: `/compare`

**Método**: `POST`

**Content-Type**: `application/json`

**Cuerpo de la Solicitud**:

```json
{
  "sample1": {
    "hu_moment_1": 0.123456,
    "hu_moment_2": 0.234567,
    "contour_area": 12345.67,
    "striation_density": 0.89
  },
  "sample2": {
    "hu_moment_1": 0.123789,
    "hu_moment_2": 0.234890,
    "contour_area": 12350.45,
    "striation_density": 0.91
  },
  "weights": {
    "hu_moment_1": 2.0,
    "striation_density": 1.5
  },
  "threshold": 0.85
}
```

**Parámetros**:

| Nombre | Tipo | Descripción | Requerido |
|--------|------|-------------|----------|
| sample1 | Object | Características de la primera muestra | Sí |
| sample2 | Object | Características de la segunda muestra | Sí |
| weights | Object | Pesos para cada característica (opcional) | No |
| threshold | Number | Umbral de coincidencia (0.0-1.0) | No |

**Respuesta Exitosa**:

- **Código**: 200 OK
- **Content-Type**: `application/json`
- **Cuerpo**:

```json
{
  "similarity": 0.92,
  "match": true,
  "confidence": 0.87,
  "feature_weights": {
    "hu_moment_1": 2.0,
    "striation_density": 1.5
  },
  "diff_per_feature": {
    "hu_moment_1": 0.000333,
    "hu_moment_2": 0.000323,
    "contour_area": 4.78,
    "striation_density": 0.02
  },
  "areas_of_interest": []
}
```

**Respuestas de Error**:

- **Código**: 400 Bad Request
  - Cuando las muestras están vacías o son inválidas
- **Código**: 500 Internal Server Error
  - Cuando ocurre un error durante la comparación

## Códigos de Estado

| Código | Descripción |
|--------|-------------|
| 200 | OK - La solicitud se completó correctamente |
| 400 | Bad Request - La solicitud contiene datos inválidos |
| 405 | Method Not Allowed - Método HTTP no soportado |
| 415 | Unsupported Media Type - Formato de archivo no soportado |
| 500 | Internal Server Error - Error en el servidor |

## Tipos de Datos

### BallisticAnalysis

| Campo | Tipo | Descripción |
|-------|------|-------------|
| features | Object | Características extraídas de la imagen |
| chroma_data | Object | Datos de análisis de color |
| metadata | Object | Metadatos del análisis |

### AnalysisMetadata

| Campo | Tipo | Descripción |
|-------|------|-------------|
| timestamp | String | Fecha y hora del análisis en formato ISO 8601 |
| image_hash | String | Hash único de la imagen procesada |
| processor_version | String | Versión del procesador de imágenes |
| python_features_used | Boolean | Indica si se utilizaron características extraídas con Python |
| confidence | Number | Nivel de confianza del análisis (0.0-1.0) |
| filename | String | Nombre original del archivo procesado |
| content_type | String | Tipo MIME del archivo (ej. image/jpeg, image/png) |
| file_size | Number | Tamaño del archivo en bytes |

### ComparisonResult

| Campo | Tipo | Descripción |
|-------|------|-------------|
| similarity | Number | Puntuación de similitud (0.0-1.0) |
| match | Boolean | Indica si las muestras coinciden según el umbral |
| confidence | Number | Confianza en el resultado (0.0-1.0) |
| feature_weights | Object | Pesos utilizados para cada característica |
| diff_per_feature | Object | Diferencia por característica |
| areas_of_interest | Array | Características con diferencias significativas |

## Notas de Uso

- Las imágenes deben estar en formato PNG o JPEG
- El tamaño máximo de archivo recomendado es 10MB
- Para comparaciones óptimas, utilice imágenes procesadas con la misma configuración
- Los pesos permiten enfatizar características específicas durante la comparación