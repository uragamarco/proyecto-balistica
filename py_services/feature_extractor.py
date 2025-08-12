import cv2
import numpy as np
import json
import sys
from skimage.feature import local_binary_pattern

def calculate_hu_moments(image_path):
    """Calcula los 7 momentos de Hu para la imagen preprocesada"""
    img = cv2.imread(image_path, cv2.IMREAD_GRAYSCALE)
    _, binary = cv2.threshold(img, 128, 255, cv2.THRESH_BINARY)
    moments = cv2.moments(binary)
    hu_moments = cv2.HuMoments(moments).flatten()
    return [float(m) for m in hu_moments]

def detect_striations(image_path):
    """Analiza patrones de estrías usando FFT y LBP"""
    img = cv2.imread(image_path, cv2.IMREAD_GRAYSCALE)
    
    # Análisis de frecuencia
    f = np.fft.fft2(img)
    fshift = np.fft.fftshift(f)
    magnitude_spectrum = 20 * np.log(np.abs(fshift))
    
    # Local Binary Patterns
    lbp = local_binary_pattern(img, 8, 1, method='uniform')
    hist, _ = np.histogram(lbp, bins=10, range=(0, 10))
    
    return [float(x) for x in hist]

def main():
    if len(sys.argv) != 2:
        print(json.dumps({"error": "Se requiere exactamente 1 argumento: path de imagen"}))
        sys.exit(1)
    
    try:
        response = {
            "hu_moments": calculate_hu_moments(sys.argv[1]),
            "striations": detect_striations(sys.argv[1]),
            "contour_area": 0.0,  # Placeholder - implementar según necesidad
            "contour_len": 0.0    # Placeholder
        }
        print(json.dumps(response))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)

if __name__ == "__main__":
    main()