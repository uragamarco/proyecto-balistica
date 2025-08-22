package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ClassificationRepository maneja las operaciones de almacenamiento para clasificaciones balísticas
type ClassificationRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewClassificationRepository crea un nuevo repositorio de clasificaciones
func NewClassificationRepository(db *Database, logger *zap.Logger) *ClassificationRepository {
	return &ClassificationRepository{
		db:     db,
		logger: logger,
	}
}

// SaveClassification guarda una clasificación balística en la base de datos
func (r *ClassificationRepository) SaveClassification(analysisID, weaponType, caliber string, confidence float64, classificationData map[string]interface{}) (*BallisticClassification, error) {
	id := uuid.New().String()
	now := time.Now()

	// Serializar datos de clasificación
	var classificationJSON []byte
	var err error
	if classificationData != nil {
		classificationJSON, err = json.Marshal(classificationData)
		if err != nil {
			return nil, fmt.Errorf("error al serializar datos de clasificación: %w", err)
		}
	}

	// Insertar en la base de datos
	query := `
		INSERT INTO ballistic_classifications (id, analysis_id, weapon_type, caliber, confidence, classification_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.db.Exec(query, id, analysisID, weaponType, caliber, confidence, string(classificationJSON), now)
	if err != nil {
		return nil, fmt.Errorf("error al guardar clasificación: %w", err)
	}

	r.logger.Info("Clasificación guardada exitosamente",
		zap.String("id", id),
		zap.String("analysis_id", analysisID),
		zap.String("weapon_type", weaponType),
		zap.String("caliber", caliber),
		zap.Float64("confidence", confidence))

	return &BallisticClassification{
		ID:                 id,
		AnalysisID:         analysisID,
		WeaponType:         weaponType,
		Caliber:            caliber,
		Confidence:         confidence,
		ClassificationData: classificationData,
		CreatedAt:          now,
	}, nil
}

// GetClassification recupera una clasificación por su ID
func (r *ClassificationRepository) GetClassification(id string) (*BallisticClassification, error) {
	query := `
		SELECT id, analysis_id, weapon_type, caliber, confidence, classification_data, created_at
		FROM ballistic_classifications
		WHERE id = ?
	`

	row := r.db.db.QueryRow(query, id)

	var classification BallisticClassification
	var weaponType, caliber sql.NullString
	var classificationJSON sql.NullString

	err := row.Scan(
		&classification.ID,
		&classification.AnalysisID,
		&weaponType,
		&caliber,
		&classification.Confidence,
		&classificationJSON,
		&classification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("clasificación no encontrada: %s", id)
		}
		return nil, fmt.Errorf("error al recuperar clasificación: %w", err)
	}

	// Asignar valores opcionales
	if weaponType.Valid {
		classification.WeaponType = weaponType.String
	}
	if caliber.Valid {
		classification.Caliber = caliber.String
	}

	// Deserializar datos de clasificación
	if classificationJSON.Valid && classificationJSON.String != "" {
		var classificationData map[string]interface{}
		if err := json.Unmarshal([]byte(classificationJSON.String), &classificationData); err != nil {
			return nil, fmt.Errorf("error al deserializar datos de clasificación: %w", err)
		}
		classification.ClassificationData = classificationData
	}

	return &classification, nil
}

// GetClassificationsByAnalysis recupera todas las clasificaciones para un análisis específico
func (r *ClassificationRepository) GetClassificationsByAnalysis(analysisID string) ([]*BallisticClassification, error) {
	query := `
		SELECT id, analysis_id, weapon_type, caliber, confidence, classification_data, created_at
		FROM ballistic_classifications
		WHERE analysis_id = ?
		ORDER BY confidence DESC, created_at DESC
	`

	rows, err := r.db.db.Query(query, analysisID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar clasificaciones: %w", err)
	}
	defer rows.Close()

	return r.scanClassifications(rows)
}

// GetClassificationsByWeaponType recupera clasificaciones por tipo de arma
func (r *ClassificationRepository) GetClassificationsByWeaponType(weaponType string, limit, offset int) ([]*BallisticClassification, error) {
	query := `
		SELECT id, analysis_id, weapon_type, caliber, confidence, classification_data, created_at
		FROM ballistic_classifications
		WHERE weapon_type = ?
		ORDER BY confidence DESC, created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, weaponType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar clasificaciones por tipo de arma: %w", err)
	}
	defer rows.Close()

	return r.scanClassifications(rows)
}

// GetClassificationsByCaliber recupera clasificaciones por calibre
func (r *ClassificationRepository) GetClassificationsByCaliber(caliber string, limit, offset int) ([]*BallisticClassification, error) {
	query := `
		SELECT id, analysis_id, weapon_type, caliber, confidence, classification_data, created_at
		FROM ballistic_classifications
		WHERE caliber = ?
		ORDER BY confidence DESC, created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, caliber, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar clasificaciones por calibre: %w", err)
	}
	defer rows.Close()

	return r.scanClassifications(rows)
}

// GetAllClassifications recupera todas las clasificaciones con paginación
func (r *ClassificationRepository) GetAllClassifications(limit, offset int) ([]*BallisticClassification, error) {
	query := `
		SELECT id, analysis_id, weapon_type, caliber, confidence, classification_data, created_at
		FROM ballistic_classifications
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar clasificaciones: %w", err)
	}
	defer rows.Close()

	return r.scanClassifications(rows)
}

// scanClassifications es un método auxiliar para escanear filas de clasificaciones
func (r *ClassificationRepository) scanClassifications(rows *sql.Rows) ([]*BallisticClassification, error) {
	var classifications []*BallisticClassification

	for rows.Next() {
		var classification BallisticClassification
		var weaponType, caliber sql.NullString
		var classificationJSON sql.NullString

		err := rows.Scan(
			&classification.ID,
			&classification.AnalysisID,
			&weaponType,
			&caliber,
			&classification.Confidence,
			&classificationJSON,
			&classification.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error al escanear clasificación: %w", err)
		}

		// Asignar valores opcionales
		if weaponType.Valid {
			classification.WeaponType = weaponType.String
		}
		if caliber.Valid {
			classification.Caliber = caliber.String
		}

		// Deserializar datos de clasificación
		if classificationJSON.Valid && classificationJSON.String != "" {
			var classificationData map[string]interface{}
			if err := json.Unmarshal([]byte(classificationJSON.String), &classificationData); err != nil {
				r.logger.Warn("Error al deserializar datos de clasificación", zap.String("id", classification.ID), zap.Error(err))
			} else {
				classification.ClassificationData = classificationData
			}
		}

		classifications = append(classifications, &classification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar clasificaciones: %w", err)
	}

	return classifications, nil
}

// DeleteClassification elimina una clasificación por su ID
func (r *ClassificationRepository) DeleteClassification(id string) error {
	query := "DELETE FROM ballistic_classifications WHERE id = ?"

	result, err := r.db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar clasificación: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar eliminación: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("clasificación no encontrada: %s", id)
	}

	r.logger.Info("Clasificación eliminada exitosamente", zap.String("id", id))
	return nil
}

// GetClassificationCount obtiene el número total de clasificaciones
func (r *ClassificationRepository) GetClassificationCount() (int, error) {
	query := "SELECT COUNT(*) FROM ballistic_classifications"

	var count int
	err := r.db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error al contar clasificaciones: %w", err)
	}

	return count, nil
}

// GetClassificationStats obtiene estadísticas de clasificaciones
func (r *ClassificationRepository) GetClassificationStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Contar total de clasificaciones
	totalQuery := "SELECT COUNT(*) FROM ballistic_classifications"
	var total int
	if err := r.db.db.QueryRow(totalQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("error al contar clasificaciones totales: %w", err)
	}
	stats["total_classifications"] = total

	// Contar por tipo de arma
	weaponTypeQuery := `
		SELECT weapon_type, COUNT(*) as count
		FROM ballistic_classifications
		WHERE weapon_type IS NOT NULL AND weapon_type != ''
		GROUP BY weapon_type
		ORDER BY count DESC
	`

	weaponTypeRows, err := r.db.db.Query(weaponTypeQuery)
	if err != nil {
		return nil, fmt.Errorf("error al consultar estadísticas por tipo de arma: %w", err)
	}
	defer weaponTypeRows.Close()

	weaponTypeStats := make(map[string]int)
	for weaponTypeRows.Next() {
		var weaponType string
		var count int
		if parseErr := weaponTypeRows.Scan(&weaponType, &count); parseErr != nil {
			return nil, fmt.Errorf("error al escanear estadísticas de tipo de arma: %w", parseErr)
		}
		weaponTypeStats[weaponType] = count
	}
	stats["weapon_type_distribution"] = weaponTypeStats

	// Contar por calibre
	caliberQuery := `
		SELECT caliber, COUNT(*) as count
		FROM ballistic_classifications
		WHERE caliber IS NOT NULL AND caliber != ''
		GROUP BY caliber
		ORDER BY count DESC
	`

	caliberRows, err := r.db.db.Query(caliberQuery)
	if err != nil {
		return nil, fmt.Errorf("error al consultar estadísticas por calibre: %w", err)
	}
	defer caliberRows.Close()

	caliberStats := make(map[string]int)
	for caliberRows.Next() {
		var caliber string
		var count int
		if err := caliberRows.Scan(&caliber, &count); err != nil {
			return nil, fmt.Errorf("error al escanear estadísticas de calibre: %w", err)
		}
		caliberStats[caliber] = count
	}
	stats["caliber_distribution"] = caliberStats

	// Promedio de confianza
	avgConfidenceQuery := "SELECT AVG(confidence) FROM ballistic_classifications"
	var avgConfidence sql.NullFloat64
	if err := r.db.db.QueryRow(avgConfidenceQuery).Scan(&avgConfidence); err != nil {
		return nil, fmt.Errorf("error al calcular promedio de confianza: %w", err)
	}
	if avgConfidence.Valid {
		stats["average_confidence"] = avgConfidence.Float64
	} else {
		stats["average_confidence"] = 0.0
	}

	return stats, nil
}