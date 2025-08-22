package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uragamarco/proyecto-balistica/internal/models"
	"go.uber.org/zap"
)

// AnalysisRepository maneja las operaciones de almacenamiento para análisis balísticos
type AnalysisRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewAnalysisRepository crea un nuevo repositorio de análisis
func NewAnalysisRepository(db *Database, logger *zap.Logger) *AnalysisRepository {
	return &AnalysisRepository{
		db:     db,
		logger: logger,
	}
}

// SaveAnalysis guarda un análisis balístico en la base de datos
func (r *AnalysisRepository) SaveAnalysis(imagePath string, features map[string]float64, metadata *models.AnalysisMetadata) (*BallisticAnalysis, error) {
	id := uuid.New().String()
	now := time.Now()

	// Serializar características
	featuresJSON, err := json.Marshal(features)
	if err != nil {
		return nil, fmt.Errorf("error al serializar características: %w", err)
	}

	// Serializar metadata
	var metadataJSON []byte
	if metadata != nil {
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("error al serializar metadata: %w", err)
		}
	}

	// Insertar en la base de datos
	query := `
		INSERT INTO ballistic_analysis (id, image_path, features, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.db.Exec(query, id, imagePath, string(featuresJSON), string(metadataJSON), now, now)
	if err != nil {
		return nil, fmt.Errorf("error al guardar análisis: %w", err)
	}

	r.logger.Info("Análisis guardado exitosamente",
		zap.String("id", id),
		zap.String("image_path", imagePath),
		zap.Int("features_count", len(features)))

	return &BallisticAnalysis{
		ID:        id,
		ImagePath: imagePath,
		Features:  features,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetAnalysis recupera un análisis por su ID
func (r *AnalysisRepository) GetAnalysis(id string) (*BallisticAnalysis, error) {
	query := `
		SELECT id, image_path, features, metadata, created_at, updated_at
		FROM ballistic_analysis
		WHERE id = ?
	`

	row := r.db.db.QueryRow(query, id)

	var analysis BallisticAnalysis
	var featuresJSON, metadataJSON sql.NullString

	err := row.Scan(
		&analysis.ID,
		&analysis.ImagePath,
		&featuresJSON,
		&metadataJSON,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("análisis no encontrado: %s", id)
		}
		return nil, fmt.Errorf("error al recuperar análisis: %w", err)
	}

	// Deserializar características
	if featuresJSON.Valid {
		if err := json.Unmarshal([]byte(featuresJSON.String), &analysis.Features); err != nil {
			return nil, fmt.Errorf("error al deserializar características: %w", err)
		}
	}

	// Deserializar metadata
	if metadataJSON.Valid && metadataJSON.String != "" {
		var metadata models.AnalysisMetadata
		if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err != nil {
			return nil, fmt.Errorf("error al deserializar metadata: %w", err)
		}
		analysis.Metadata = &metadata
	}

	return &analysis, nil
}

// GetAllAnalysis recupera todos los análisis con paginación
func (r *AnalysisRepository) GetAllAnalysis(limit, offset int) ([]*BallisticAnalysis, error) {
	query := `
		SELECT id, image_path, features, metadata, created_at, updated_at
		FROM ballistic_analysis
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al consultar análisis: %w", err)
	}
	defer rows.Close()

	var analyses []*BallisticAnalysis

	for rows.Next() {
		var analysis BallisticAnalysis
		var featuresJSON, metadataJSON sql.NullString

		err := rows.Scan(
			&analysis.ID,
			&analysis.ImagePath,
			&featuresJSON,
			&metadataJSON,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error al escanear análisis: %w", err)
		}

		// Deserializar características
		if featuresJSON.Valid {
			if err := json.Unmarshal([]byte(featuresJSON.String), &analysis.Features); err != nil {
				r.logger.Warn("Error al deserializar características", zap.String("id", analysis.ID), zap.Error(err))
				continue
			}
		}

		// Deserializar metadata
		if metadataJSON.Valid && metadataJSON.String != "" {
			var metadata models.AnalysisMetadata
			if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err != nil {
				r.logger.Warn("Error al deserializar metadata", zap.String("id", analysis.ID), zap.Error(err))
			} else {
				analysis.Metadata = &metadata
			}
		}

		analyses = append(analyses, &analysis)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar análisis: %w", err)
	}

	return analyses, nil
}

// SearchAnalysisByImagePath busca análisis por ruta de imagen
func (r *AnalysisRepository) SearchAnalysisByImagePath(imagePath string) ([]*BallisticAnalysis, error) {
	query := `
		SELECT id, image_path, features, metadata, created_at, updated_at
		FROM ballistic_analysis
		WHERE image_path LIKE ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.db.Query(query, "%"+imagePath+"%")
	if err != nil {
		return nil, fmt.Errorf("error al buscar análisis: %w", err)
	}
	defer rows.Close()

	var analyses []*BallisticAnalysis

	for rows.Next() {
		var analysis BallisticAnalysis
		var featuresJSON, metadataJSON sql.NullString

		err := rows.Scan(
			&analysis.ID,
			&analysis.ImagePath,
			&featuresJSON,
			&metadataJSON,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error al escanear análisis: %w", err)
		}

		// Deserializar características
		if featuresJSON.Valid {
			if err := json.Unmarshal([]byte(featuresJSON.String), &analysis.Features); err != nil {
				r.logger.Warn("Error al deserializar características", zap.String("id", analysis.ID), zap.Error(err))
				continue
			}
		}

		// Deserializar metadata
		if metadataJSON.Valid && metadataJSON.String != "" {
			var metadata models.AnalysisMetadata
			if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err != nil {
				r.logger.Warn("Error al deserializar metadata", zap.String("id", analysis.ID), zap.Error(err))
			} else {
				analysis.Metadata = &metadata
			}
		}

		analyses = append(analyses, &analysis)
	}

	return analyses, nil
}

// DeleteAnalysis elimina un análisis por su ID
func (r *AnalysisRepository) DeleteAnalysis(id string) error {
	query := "DELETE FROM ballistic_analysis WHERE id = ?"

	result, err := r.db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar análisis: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar eliminación: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("análisis no encontrado: %s", id)
	}

	r.logger.Info("Análisis eliminado exitosamente", zap.String("id", id))
	return nil
}

// GetAnalysisCount obtiene el número total de análisis
func (r *AnalysisRepository) GetAnalysisCount() (int, error) {
	query := "SELECT COUNT(*) FROM ballistic_analysis"

	var count int
	err := r.db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error al contar análisis: %w", err)
	}

	return count, nil
}