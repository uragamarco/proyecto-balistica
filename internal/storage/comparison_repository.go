package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ComparisonRepository maneja las operaciones de almacenamiento para comparaciones balísticas
type ComparisonRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewComparisonRepository crea un nuevo repositorio de comparaciones
func NewComparisonRepository(db *Database, logger *zap.Logger) *ComparisonRepository {
	return &ComparisonRepository{
		db:     db,
		logger: logger,
	}
}

// SaveComparison guarda una comparación balística en la base de datos
func (r *ComparisonRepository) SaveComparison(sample1ID, sample2ID string, similarity, confidence float64, matchResult bool, comparisonData map[string]interface{}) (*BallisticComparison, error) {
	id := uuid.New().String()
	now := time.Now()

	// Serializar datos de comparación
	var comparisonJSON []byte
	var err error
	if comparisonData != nil {
		comparisonJSON, err = json.Marshal(comparisonData)
		if err != nil {
			return nil, fmt.Errorf("error al serializar datos de comparación: %w", err)
		}
	}

	// Insertar en la base de datos
	query := `
		INSERT INTO ballistic_comparisons (id, sample1_id, sample2_id, similarity, confidence, match_result, comparison_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.db.Exec(query, id, sample1ID, sample2ID, similarity, confidence, matchResult, string(comparisonJSON), now)
	if err != nil {
		return nil, fmt.Errorf("error al guardar comparación: %w", err)
	}

	r.logger.Info("Comparación guardada exitosamente",
		zap.String("id", id),
		zap.String("sample1_id", sample1ID),
		zap.String("sample2_id", sample2ID),
		zap.Float64("similarity", similarity),
		zap.Bool("match", matchResult))

	return &BallisticComparison{
		ID:             id,
		Sample1ID:      sample1ID,
		Sample2ID:      sample2ID,
		Similarity:     similarity,
		Confidence:     confidence,
		MatchResult:    matchResult,
		ComparisonData: comparisonData,
		CreatedAt:      now,
	}, nil
}

// GetComparison recupera una comparación por su ID
func (r *ComparisonRepository) GetComparison(id string) (*BallisticComparison, error) {
	query := `
		SELECT id, sample1_id, sample2_id, similarity, confidence, match_result, comparison_data, created_at
		FROM ballistic_comparisons
		WHERE id = ?
	`

	row := r.db.db.QueryRow(query, id)

	var comparison BallisticComparison
	var comparisonJSON sql.NullString

	err := row.Scan(
		&comparison.ID,
		&comparison.Sample1ID,
		&comparison.Sample2ID,
		&comparison.Similarity,
		&comparison.Confidence,
		&comparison.MatchResult,
		&comparisonJSON,
		&comparison.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comparación no encontrada: %s", id)
		}
		return nil, fmt.Errorf("error al recuperar comparación: %w", err)
	}

	// Deserializar datos de comparación
	if comparisonJSON.Valid && comparisonJSON.String != "" {
		var comparisonData map[string]interface{}
		if err := json.Unmarshal([]byte(comparisonJSON.String), &comparisonData); err != nil {
			return nil, fmt.Errorf("error al deserializar datos de comparación: %w", err)
		}
		comparison.ComparisonData = comparisonData
	}

	return &comparison, nil
}

// GetComparisonsBySample recupera todas las comparaciones que involucran una muestra específica
func (r *ComparisonRepository) GetComparisonsBySample(sampleID string, limit, offset int) ([]*BallisticComparison, error) {
	query := `
		SELECT id, sample1_id, sample2_id, similarity, confidence, match_result, comparison_data, created_at
		FROM ballistic_comparisons
		WHERE sample1_id = ? OR sample2_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, sampleID, sampleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comparaciones: %w", err)
	}
	defer rows.Close()

	return r.scanComparisons(rows)
}

// GetAllComparisons recupera todas las comparaciones con paginación
func (r *ComparisonRepository) GetAllComparisons(limit, offset int) ([]*BallisticComparison, error) {
	query := `
		SELECT id, sample1_id, sample2_id, similarity, confidence, match_result, comparison_data, created_at
		FROM ballistic_comparisons
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comparaciones: %w", err)
	}
	defer rows.Close()

	return r.scanComparisons(rows)
}

// GetMatchingComparisons recupera comparaciones que resultaron en coincidencias
func (r *ComparisonRepository) GetMatchingComparisons(limit, offset int) ([]*BallisticComparison, error) {
	query := `
		SELECT id, sample1_id, sample2_id, similarity, confidence, match_result, comparison_data, created_at
		FROM ballistic_comparisons
		WHERE match_result = true
		ORDER BY similarity DESC, created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comparaciones coincidentes: %w", err)
	}
	defer rows.Close()

	return r.scanComparisons(rows)
}

// GetComparisonsByDateRange recupera comparaciones en un rango de fechas
func (r *ComparisonRepository) GetComparisonsByDateRange(startDate, endDate time.Time, limit, offset int) ([]*BallisticComparison, error) {
	query := `
		SELECT id, sample1_id, sample2_id, similarity, confidence, match_result, comparison_data, created_at
		FROM ballistic_comparisons
		WHERE created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comparaciones por fecha: %w", err)
	}
	defer rows.Close()

	return r.scanComparisons(rows)
}

// scanComparisons es un método auxiliar para escanear filas de comparaciones
func (r *ComparisonRepository) scanComparisons(rows *sql.Rows) ([]*BallisticComparison, error) {
	var comparisons []*BallisticComparison

	for rows.Next() {
		var comparison BallisticComparison
		var comparisonJSON sql.NullString

		err := rows.Scan(
			&comparison.ID,
			&comparison.Sample1ID,
			&comparison.Sample2ID,
			&comparison.Similarity,
			&comparison.Confidence,
			&comparison.MatchResult,
			&comparisonJSON,
			&comparison.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error al escanear comparación: %w", err)
		}

		// Deserializar datos de comparación
		if comparisonJSON.Valid && comparisonJSON.String != "" {
			var comparisonData map[string]interface{}
			if err := json.Unmarshal([]byte(comparisonJSON.String), &comparisonData); err != nil {
				r.logger.Warn("Error al deserializar datos de comparación", zap.String("id", comparison.ID), zap.Error(err))
			} else {
				comparison.ComparisonData = comparisonData
			}
		}

		comparisons = append(comparisons, &comparison)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar comparaciones: %w", err)
	}

	return comparisons, nil
}

// DeleteComparison elimina una comparación por su ID
func (r *ComparisonRepository) DeleteComparison(id string) error {
	query := "DELETE FROM ballistic_comparisons WHERE id = ?"

	result, err := r.db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar comparación: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar eliminación: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comparación no encontrada: %s", id)
	}

	r.logger.Info("Comparación eliminada exitosamente", zap.String("id", id))
	return nil
}

// GetComparisonCount obtiene el número total de comparaciones
func (r *ComparisonRepository) GetComparisonCount() (int, error) {
	query := "SELECT COUNT(*) FROM ballistic_comparisons"

	var count int
	err := r.db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error al contar comparaciones: %w", err)
	}

	return count, nil
}

// GetComparisonStats obtiene estadísticas de comparaciones
func (r *ComparisonRepository) GetComparisonStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Contar total de comparaciones
	totalQuery := "SELECT COUNT(*) FROM ballistic_comparisons"
	var total int
	if err := r.db.db.QueryRow(totalQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("error al contar comparaciones totales: %w", err)
	}
	stats["total_comparisons"] = total

	// Contar coincidencias
	matchQuery := "SELECT COUNT(*) FROM ballistic_comparisons WHERE match_result = true"
	var matches int
	if err := r.db.db.QueryRow(matchQuery).Scan(&matches); err != nil {
		return nil, fmt.Errorf("error al contar coincidencias: %w", err)
	}
	stats["total_matches"] = matches

	// Calcular porcentaje de coincidencias
	if total > 0 {
		stats["match_percentage"] = float64(matches) / float64(total) * 100
	} else {
		stats["match_percentage"] = 0.0
	}

	// Promedio de similitud
	avgQuery := "SELECT AVG(similarity) FROM ballistic_comparisons"
	var avgSimilarity sql.NullFloat64
	if err := r.db.db.QueryRow(avgQuery).Scan(&avgSimilarity); err != nil {
		return nil, fmt.Errorf("error al calcular promedio de similitud: %w", err)
	}
	if avgSimilarity.Valid {
		stats["average_similarity"] = avgSimilarity.Float64
	} else {
		stats["average_similarity"] = 0.0
	}

	// Promedio de confianza
	avgConfidenceQuery := "SELECT AVG(confidence) FROM ballistic_comparisons"
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