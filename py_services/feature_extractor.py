import cv2
import numpy as np
import logging
import os
from flask import Flask, request, jsonify

app = Flask(__name__)

# Configuración avanzada de logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("feature_extractor.log"),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger('FeatureExtractor')

class FeatureExtractionError(Exception):
    pass

def calculate_hu_moments(image_data: bytes) -> list:
    try:
        nparr = np.frombuffer(image_data, np.uint8)
        img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
        if img is None:
            raise FeatureExtractionError("Formato de imagen no soportado")
        
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
        blur = cv2.GaussianBlur(gray, (5, 5), 0)
        
        thresh = cv2.adaptiveThreshold(
            blur, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
            cv2.THRESH_BINARY_INV, 11, 2
        )
        
        contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        if not contours:
            logger.warning("No se encontraron contornos")
            return [0.0] * 7
        
        largest_contour = max(contours, key=cv2.contourArea)
        mask = np.zeros_like(gray)
        cv2.drawContours(mask, [largest_contour], -1, 255, thickness=cv2.FILLED)
        
        masked_thresh = cv2.bitwise_and(thresh, thresh, mask=mask)
        moments = cv2.moments(masked_thresh)
        hu_moments = cv2.HuMoments(moments).flatten()
        
        hu_moments = np.where(
            np.abs(hu_moments) > 1e-9,
            -np.sign(hu_moments) * np.log10(np.abs(hu_moments)),
            0
        )
        
        return hu_moments.tolist()
        
    except Exception as e:
        logger.exception("Error en cálculo de momentos de Hu")
        raise FeatureExtractionError(f"Error de procesamiento: {str(e)}")

@app.route('/extract', methods=['POST'])
def extract_features_endpoint():
    try:
        if 'image' not in request.files:
            return jsonify({"error": "No se proporcionó imagen"}), 400
            
        file = request.files['image']
        image_data = file.read()
        
        hu_moments = calculate_hu_moments(image_data)
        
        return jsonify({
            "hu_moments": hu_moments,
            "status": "success"
        })
        
    except FeatureExtractionError as e:
        return jsonify({"error": str(e)}), 500
    except Exception as e:
        logger.exception("Error inesperado")
        return jsonify({"error": "Error interno del servidor"}), 500

@app.route('/health')
def health_check():
    return jsonify({
        "status": "healthy",
        "service": "feature-extractor",
        "version": "1.0.0"
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
    