# Manejo de Errores en Proyecto Balística

## Introducción

Este documento describe el sistema de manejo de errores implementado en el Proyecto Balística. El sistema está diseñado para proporcionar respuestas de error consistentes, facilitar la depuración y mejorar la experiencia del usuario.

## Estructura de Errores

### APIError

La estructura principal para el manejo de errores es `APIError`, definida en `internal/api/handlers.go`:

```go
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
```

Donde:
- `Code`: Código HTTP del error (ej. 400, 404, 500)
- `Message`: Mensaje descriptivo del error
- `Details`: Detalles adicionales, generalmente el mensaje de error original

## Funciones de Respuesta

### respondWithError

Utilizada para enviar respuestas de error en formato JSON:

```go
func (h *Handlers) respondWithError(w http.ResponseWriter, apiErr APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	
	// Registrar el error en los logs
	h.Logger.Error("API Error",
		zap.Int("code", apiErr.Code),
		zap.String("message", apiErr.Message),
		zap.String("details", apiErr.Details))
	
	// Enviar respuesta JSON
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		h.Logger.Error("Error al serializar respuesta de error", zap.Error(err))
	}
}
```

### respondWithJSON

Utilizada para enviar respuestas exitosas en formato JSON:

```go
func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			h.Logger.Error("Error al serializar respuesta JSON", zap.Error(err))
			h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al generar respuesta", err))
			return
		}
	}
}
```

## Logging

El sistema utiliza [zap](https://github.com/uber-go/zap) para logging estructurado. Cada error incluye:

- Código HTTP
- Mensaje de error
- Detalles adicionales
- Información contextual (cuando aplica)

## Integración Python-Go

La integración entre Python y Go incluye manejo de errores mejorado:

- Verificación de existencia de archivos
- Timeouts para operaciones de extracción de características
- Parsing de respuestas de error desde Python
- Logging detallado de errores

## Ejemplos de Uso

### Validación de Método HTTP

```go
if r.Method != http.MethodPost {
	h.respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "Método no permitido", nil))
	return
}
```

### Validación de Entrada

```go
if len(comparisonRequest.Sample1) == 0 || len(comparisonRequest.Sample2) == 0 {
	h.respondWithError(w, NewAPIError(http.StatusBadRequest, "Muestras vacías", errors.New("ambas muestras deben contener datos")))
	return
}
```

### Errores de Procesamiento

```go
processedImg, err := h.imageProcessor.Process(img)
if err != nil {
	h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al procesar la imagen", err))
	return
}
```

## Mejores Prácticas

1. **Mensajes Descriptivos**: Usar mensajes claros que describan el problema
2. **Códigos HTTP Apropiados**: Utilizar el código HTTP que mejor describa el tipo de error
3. **Logging Consistente**: Registrar todos los errores con información contextual
4. **Manejo de Errores Anidados**: Propagar errores con contexto adicional
5. **Respuestas al Cliente**: Proporcionar información útil sin exponer detalles sensibles

## Conclusión

El sistema de manejo de errores implementado proporciona una forma consistente y estructurada de manejar errores en toda la aplicación, mejorando la depuración y la experiencia del usuario.