package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
	"gocv.io/x/gocv"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Validar método y tamaño
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB

	// Obtener archivo y acción
	file, header, err := r.FormFile("image")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error al leer imagen")
		return
	}
	defer file.Close()

	action := r.FormValue("action")
	if action != "add" && action != "compare" {
		respondWithError(w, http.StatusBadRequest, "Acción inválida")
		return
	}

	// Procesar imagen
	fileBytes, _ := io.ReadAll(file)
	img, err := gocv.IMDecode(fileBytes, gocv.IMReadColor)
	if err != nil || img.Empty() {
		respondWithError(w, http.StatusBadRequest, "Imagen inválida")
		return
	}
	defer img.Close()

	// Extraer características
	processor := image_processor.New()
	kp, desc, err := processor.Process(img)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer desc.Close()

	// Manejar acción
	switch action {
	case "add":
		if err := chroma.GetService().StoreDescriptors(filepath.Base(header.Filename), desc); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error al almacenar descriptores")
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message":   fmt.Sprintf("Imagen '%s' agregada", header.Filename),
			"keypoints": fmt.Sprintf("%d", len(kp)),
		})

	case "compare":
		results, err := chroma.GetService().QuerySimilar(desc, 5)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error en la consulta")
			return
		}
		respondWithJSON(w, http.StatusOK, results)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
