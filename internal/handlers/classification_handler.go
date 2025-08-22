package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/uragamarco/proyecto-balistica/internal/services/classification"
)

// ClassificationHandler maneja las solicitudes relacionadas con clasificación balística
type ClassificationHandler struct {
	classificationService *classification.ClassificationService
	logger                *zap.Logger
}

// NewClassificationHandler crea una nueva instancia del handler de clasificación
func NewClassificationHandler(
	classificationService *classification.ClassificationService,
	logger *zap.Logger,
) *ClassificationHandler {
	return &ClassificationHandler{
		classificationService: classificationService,
		logger:                logger,
	}
}

// ClassifyBallisticRequest estructura de solicitud para clasificación
type ClassifyBallisticRequest struct {
	AnalysisID string             `json:"analysis_id"`
	Features   map[string]float64 `json:"features"`
}

// ClassifyBallisticResponse estructura de respuesta para clasificación
type ClassifyBallisticResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ClassificationHistoryResponse estructura de respuesta para historial
type ClassificationHistoryResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    []ClassificationSummary `json:"data,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// ClassificationSummary resumen de clasificación para listados
type ClassificationSummary struct {
	ID           string  `json:"id"`
	AnalysisID   string  `json:"analysis_id"`
	WeaponType   string  `json:"weapon_type"`
	Caliber      string  `json:"caliber"`
	Confidence   float64 `json:"confidence"`
	CreatedAt    string  `json:"created_at"`
}

// SearchResponse estructura de respuesta para búsquedas
type SearchResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    []ClassificationSummary `json:"data,omitempty"`
	Count   int                     `json:"count"`
	Error   string                  `json:"error,omitempty"`
}

// ClassifyBallistic maneja la clasificación de una muestra balística
// @Summary Clasificar muestra balística
// @Description Realiza clasificación completa de tipo de arma y calibre basándose en características extraídas
// @Tags classification
// @Accept json
// @Produce json
// @Param request body ClassifyBallisticRequest true "Datos de clasificación"
// @Success 200 {object} ClassifyBallisticResponse
// @Failure 400 {object} ClassifyBallisticResponse
// @Failure 500 {object} ClassifyBallisticResponse
// @Router /api/classification/classify [post]
func (ch *ClassificationHandler) ClassifyBallistic(w http.ResponseWriter, r *http.Request) {
	ch.logger.Info("Solicitud de clasificación balística recibida")

	var req ClassifyBallisticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Error decodificando solicitud", zap.Error(err))
		ch.sendErrorResponse(w, http.StatusBadRequest, "Formato de solicitud inválido", err.Error())
		return
	}

	// Validar solicitud
	if req.AnalysisID == "" {
		ch.sendErrorResponse(w, http.StatusBadRequest, "ID de análisis requerido", "analysis_id no puede estar vacío")
		return
	}

	if len(req.Features) == 0 {
		ch.sendErrorResponse(w, http.StatusBadRequest, "Características requeridas", "features no puede estar vacío")
		return
	}

	// Realizar clasificación
	result, err := ch.classificationService.ClassifyBallistic(r.Context(), req.AnalysisID, req.Features)
	if err != nil {
		ch.logger.Error("Error en clasificación balística",
			zap.String("analysis_id", req.AnalysisID),
			zap.Error(err))
		ch.sendErrorResponse(w, http.StatusInternalServerError, "Error en clasificación", err.Error())
		return
	}

	// Enviar respuesta exitosa
	response := ClassifyBallisticResponse{
		Success: true,
		Message: "Clasificación completada exitosamente",
		Data:    result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Clasificación balística completada",
		zap.String("analysis_id", req.AnalysisID),
		zap.String("weapon_type", result.WeaponType.WeaponType),
		zap.String("caliber", result.Caliber.Caliber))
}

// GetClassificationHistory obtiene el historial de clasificaciones
// @Summary Obtener historial de clasificaciones
// @Description Obtiene una lista de clasificaciones realizadas previamente
// @Tags classification
// @Produce json
// @Param limit query int false "Límite de resultados" default(50)
// @Success 200 {object} ClassificationHistoryResponse
// @Failure 500 {object} ClassificationHistoryResponse
// @Router /api/classification/history [get]
func (ch *ClassificationHandler) GetClassificationHistory(w http.ResponseWriter, r *http.Request) {
	ch.logger.Info("Solicitud de historial de clasificaciones")

	// Obtener parámetro de límite
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // valor por defecto
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Obtener historial
	classifications, err := ch.classificationService.GetClassificationHistory(r.Context(), limit)
	if err != nil {
		ch.logger.Error("Error obteniendo historial", zap.Error(err))
		ch.sendHistoryErrorResponse(w, http.StatusInternalServerError, "Error obteniendo historial", err.Error())
		return
	}

	// Convertir a resumen
	summaries := make([]ClassificationSummary, len(classifications))
	for i, classification := range classifications {
		summaries[i] = ClassificationSummary{
			ID:         classification.ID,
			AnalysisID: classification.AnalysisID,
			WeaponType: classification.WeaponType,
			Caliber:    classification.Caliber,
			Confidence: classification.Confidence,
			CreatedAt:  classification.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// Enviar respuesta
	response := ClassificationHistoryResponse{
		Success: true,
		Message: "Historial obtenido exitosamente",
		Data:    summaries,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Historial de clasificaciones enviado",
		zap.Int("count", len(summaries)))
}

// GetClassificationByAnalysisID obtiene clasificación por ID de análisis
// @Summary Obtener clasificación por ID de análisis
// @Description Obtiene los detalles de una clasificación específica
// @Tags classification
// @Produce json
// @Param analysisId path string true "ID del análisis"
// @Success 200 {object} ClassifyBallisticResponse
// @Failure 404 {object} ClassifyBallisticResponse
// @Failure 500 {object} ClassifyBallisticResponse
// @Router /api/classification/analysis/{analysisId} [get]
func (ch *ClassificationHandler) GetClassificationByAnalysisID(w http.ResponseWriter, r *http.Request) {
	// Extraer analysisId de la URL
	path := strings.TrimPrefix(r.URL.Path, "/api/classification/analysis/")
	analysisID := strings.TrimSuffix(path, "/")

	ch.logger.Info("Solicitud de clasificación por ID",
		zap.String("analysis_id", analysisID))

	if analysisID == "" {
		ch.sendErrorResponse(w, http.StatusBadRequest, "ID de análisis requerido", "analysisId no puede estar vacío")
		return
	}

	// Obtener clasificación
	results, err := ch.classificationService.GetClassificationByAnalysisID(r.Context(), analysisID)
	if err != nil {
		ch.logger.Error("Error obteniendo clasificación",
			zap.String("analysis_id", analysisID),
			zap.Error(err))
		ch.sendErrorResponse(w, http.StatusNotFound, "Clasificación no encontrada", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ClassifyBallisticResponse{
		Success: true,
		Message: "Clasificación obtenida exitosamente",
		Data: map[string]interface{}{
			"classifications": results,
		},
	})

	ch.logger.Info("Clasificación enviada",
		zap.String("analysis_id", analysisID))
}

// SearchByWeaponType busca clasificaciones por tipo de arma
// @Summary Buscar por tipo de arma
// @Description Busca clasificaciones filtradas por tipo de arma
// @Tags classification
// @Produce json
// @Param weaponType path string true "Tipo de arma"
// @Success 200 {object} SearchResponse
// @Failure 500 {object} SearchResponse
// @Router /api/classification/search/weapon/{weaponType} [get]
func (ch *ClassificationHandler) SearchByWeaponType(w http.ResponseWriter, r *http.Request) {
	// Extraer weaponType de la URL
	path := strings.TrimPrefix(r.URL.Path, "/api/classification/search/weapon/")
	weaponType := strings.TrimSuffix(path, "/")

	ch.logger.Info("Búsqueda por tipo de arma",
		zap.String("weapon_type", weaponType))

	if weaponType == "" {
		ch.sendSearchErrorResponse(w, http.StatusBadRequest, "Tipo de arma requerido", "weaponType no puede estar vacío")
		return
	}

	// Buscar clasificaciones
	classifications, err := ch.classificationService.SearchByWeaponType(r.Context(), weaponType)
	if err != nil {
		ch.logger.Error("Error en búsqueda por tipo de arma",
			zap.String("weapon_type", weaponType),
			zap.Error(err))
		ch.sendSearchErrorResponse(w, http.StatusInternalServerError, "Error en búsqueda", err.Error())
		return
	}

	// Convertir a resumen
	summaries := make([]ClassificationSummary, len(classifications))
	for i, classification := range classifications {
		summaries[i] = ClassificationSummary{
			ID:         classification.ID,
			AnalysisID: classification.AnalysisID,
			WeaponType: classification.WeaponType,
			Caliber:    classification.Caliber,
			Confidence: classification.Confidence,
			CreatedAt:  classification.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// Enviar respuesta
	response := SearchResponse{
		Success: true,
		Message: "Búsqueda completada exitosamente",
		Data:    summaries,
		Count:   len(summaries),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Búsqueda por tipo de arma completada",
		zap.String("weapon_type", weaponType),
		zap.Int("results", len(summaries)))
}

// SearchByCaliber busca clasificaciones por calibre
// @Summary Buscar por calibre
// @Description Busca clasificaciones filtradas por calibre
// @Tags classification
// @Produce json
// @Param caliber path string true "Calibre"
// @Success 200 {object} SearchResponse
// @Failure 500 {object} SearchResponse
// @Router /api/classification/search/caliber/{caliber} [get]
func (ch *ClassificationHandler) SearchByCaliber(w http.ResponseWriter, r *http.Request) {
	// Extraer caliber de la URL
	path := strings.TrimPrefix(r.URL.Path, "/api/classification/search/caliber/")
	caliber := strings.TrimSuffix(path, "/")

	ch.logger.Info("Búsqueda por calibre",
		zap.String("caliber", caliber))

	if caliber == "" {
		ch.sendSearchErrorResponse(w, http.StatusBadRequest, "Calibre requerido", "caliber no puede estar vacío")
		return
	}

	// Buscar clasificaciones
	classifications, err := ch.classificationService.SearchByCaliber(r.Context(), caliber)
	if err != nil {
		ch.logger.Error("Error en búsqueda por calibre",
			zap.String("caliber", caliber),
			zap.Error(err))
		ch.sendSearchErrorResponse(w, http.StatusInternalServerError, "Error en búsqueda", err.Error())
		return
	}

	// Convertir a resumen
	summaries := make([]ClassificationSummary, len(classifications))
	for i, classification := range classifications {
		summaries[i] = ClassificationSummary{
			ID:         classification.ID,
			AnalysisID: classification.AnalysisID,
			WeaponType: classification.WeaponType,
			Caliber:    classification.Caliber,
			Confidence: classification.Confidence,
			CreatedAt:  classification.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// Enviar respuesta
	response := SearchResponse{
		Success: true,
		Message: "Búsqueda completada exitosamente",
		Data:    summaries,
		Count:   len(summaries),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Búsqueda por calibre completada",
		zap.String("caliber", caliber),
		zap.Int("results", len(summaries)))
}

// Métodos auxiliares para envío de respuestas de error

func (ch *ClassificationHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, errorDetail string) {
	response := ClassifyBallisticResponse{
		Success: false,
		Message: message,
		Error:   errorDetail,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (ch *ClassificationHandler) sendHistoryErrorResponse(w http.ResponseWriter, statusCode int, message, errorDetail string) {
	response := ClassificationHistoryResponse{
		Success: false,
		Message: message,
		Error:   errorDetail,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (ch *ClassificationHandler) sendSearchErrorResponse(w http.ResponseWriter, statusCode int, message, errorDetail string) {
	response := SearchResponse{
		Success: false,
		Message: message,
		Error:   errorDetail,
		Count:   0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}