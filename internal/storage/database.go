package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"go.uber.org/zap"
	_ "github.com/mattn/go-sqlite3"
)

// Database representa la conexión a la base de datos
type Database struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDatabase crea una nueva instancia de la base de datos
func NewDatabase(dbPath string, logger *zap.Logger) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar con la base de datos: %w", err)
	}

	database := &Database{
		db:     db,
		logger: logger,
	}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("error al crear las tablas: %w", err)
	}

	return database, nil
}

// Close cierra la conexión a la base de datos
func (d *Database) Close() error {
	return d.db.Close()
}

// createTables crea las tablas necesarias en la base de datos
func (d *Database) createTables() error {
	// Tabla para almacenar análisis de imágenes
	createAnalysisTable := `
	CREATE TABLE IF NOT EXISTS ballistic_analysis (
		id TEXT PRIMARY KEY,
		image_path TEXT NOT NULL,
		features TEXT NOT NULL,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Tabla para almacenar comparaciones
	createComparisonTable := `
	CREATE TABLE IF NOT EXISTS ballistic_comparisons (
		id TEXT PRIMARY KEY,
		sample1_id TEXT NOT NULL,
		sample2_id TEXT NOT NULL,
		similarity REAL NOT NULL,
		confidence REAL NOT NULL,
		match_result BOOLEAN NOT NULL,
		comparison_data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sample1_id) REFERENCES ballistic_analysis(id),
		FOREIGN KEY (sample2_id) REFERENCES ballistic_analysis(id)
	);
	`

	// Tabla para clasificaciones
	createClassificationTable := `
	CREATE TABLE IF NOT EXISTS ballistic_classifications (
		id TEXT PRIMARY KEY,
		analysis_id TEXT NOT NULL,
		weapon_type TEXT,
		caliber TEXT,
		confidence REAL,
		classification_data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (analysis_id) REFERENCES ballistic_analysis(id)
	);
	`

	// Índices para mejorar el rendimiento
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_analysis_created_at ON ballistic_analysis(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_comparison_samples ON ballistic_comparisons(sample1_id, sample2_id);",
		"CREATE INDEX IF NOT EXISTS idx_comparison_created_at ON ballistic_comparisons(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_classification_analysis ON ballistic_classifications(analysis_id);",
	}

	// Ejecutar creación de tablas
	for _, query := range []string{createAnalysisTable, createComparisonTable, createClassificationTable} {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("error al crear tabla: %w", err)
		}
	}

	// Ejecutar creación de índices
	for _, query := range createIndexes {
		if _, err := d.db.Exec(query); err != nil {
			d.logger.Warn("Error al crear índice", zap.Error(err))
		}
	}

	d.logger.Info("Tablas de base de datos creadas exitosamente")
	return nil
}

// BallisticAnalysis representa un análisis almacenado en la base de datos
type BallisticAnalysis struct {
	ID        string                 `json:"id"`
	ImagePath string                 `json:"image_path"`
	Features  map[string]float64     `json:"features"`
	Metadata  *models.AnalysisMetadata `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// BallisticComparison representa una comparación almacenada
type BallisticComparison struct {
	ID             string                 `json:"id"`
	Sample1ID      string                 `json:"sample1_id"`
	Sample2ID      string                 `json:"sample2_id"`
	Similarity     float64                `json:"similarity"`
	Confidence     float64                `json:"confidence"`
	MatchResult    bool                   `json:"match_result"`
	ComparisonData map[string]interface{} `json:"comparison_data,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// BallisticClassification representa una clasificación almacenada
type BallisticClassification struct {
	ID                 string                 `json:"id"`
	AnalysisID         string                 `json:"analysis_id"`
	WeaponType         string                 `json:"weapon_type,omitempty"`
	Caliber            string                 `json:"caliber,omitempty"`
	Confidence         float64                `json:"confidence"`
	ClassificationData map[string]interface{} `json:"classification_data,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}