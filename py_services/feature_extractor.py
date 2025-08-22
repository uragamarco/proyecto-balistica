import cv2
import numpy as np
import logging
import os
import sys
import json
import argparse
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

def detect_firing_pin_marks(gray_img):
    """Detecta marcas de percutor usando detección de círculos de Hough"""
    try:
        # Aplicar filtro Gaussiano para reducir ruido
        blurred = cv2.GaussianBlur(gray_img, (9, 9), 2)
        
        # Detectar círculos usando transformada de Hough
        circles = cv2.HoughCircles(
            blurred,
            cv2.HOUGH_GRADIENT,
            dp=1,
            minDist=30,
            param1=50,
            param2=30,
            minRadius=5,
            maxRadius=100
        )
        
        firing_pin_features = {
            "num_circular_marks": 0,
            "avg_mark_radius": 0.0,
            "mark_positions": [],
            "mark_intensities": []
        }
        
        if circles is not None:
            circles = np.round(circles[0, :]).astype("int")
            firing_pin_features["num_circular_marks"] = len(circles)
            
            radii = []
            intensities = []
            positions = []
            
            for (x, y, r) in circles:
                radii.append(float(r))
                positions.append([float(x), float(y)])
                
                # Calcular intensidad promedio en la región circular
                mask = np.zeros(gray_img.shape, dtype=np.uint8)
                cv2.circle(mask, (x, y), r, 255, -1)
                intensity = cv2.mean(gray_img, mask=mask)[0]
                intensities.append(float(intensity))
            
            if radii:
                firing_pin_features["avg_mark_radius"] = float(np.mean(radii))
            firing_pin_features["mark_positions"] = positions
            firing_pin_features["mark_intensities"] = intensities
        
        return firing_pin_features
        
    except Exception as e:
        logger.warning(f"Error en detección de marcas de percutor: {str(e)}")
        return {
            "num_circular_marks": 0,
            "avg_mark_radius": 0.0,
            "mark_positions": [],
            "mark_intensities": []
        }

def detect_striation_patterns(gray_img):
    """Detecta patrones de estriado usando análisis de gradientes direccionales"""
    try:
        # Aplicar filtro de mediana para reducir ruido
        denoised = cv2.medianBlur(gray_img, 5)
        
        # Calcular gradientes en X e Y
        grad_x = cv2.Sobel(denoised, cv2.CV_64F, 1, 0, ksize=3)
        grad_y = cv2.Sobel(denoised, cv2.CV_64F, 0, 1, ksize=3)
        
        # Calcular magnitud y dirección del gradiente
        magnitude = np.sqrt(grad_x**2 + grad_y**2)
        direction = np.arctan2(grad_y, grad_x)
        
        # Normalizar dirección a [0, π]
        direction = np.abs(direction)
        
        # Crear histograma de direcciones para detectar patrones dominantes
        hist, bins = np.histogram(direction.flatten(), bins=36, range=(0, np.pi))
        
        # Encontrar direcciones dominantes
        dominant_directions = []
        threshold = np.max(hist) * 0.3  # 30% del pico máximo
        
        for i, count in enumerate(hist):
            if count > threshold:
                angle = (bins[i] + bins[i+1]) / 2
                dominant_directions.append(float(angle * 180 / np.pi))  # Convertir a grados
        
        # Detectar líneas usando transformada de Hough
        edges = cv2.Canny(denoised, 50, 150)
        lines = cv2.HoughLinesP(
            edges,
            rho=1,
            theta=np.pi/180,
            threshold=50,
            minLineLength=30,
            maxLineGap=10
        )
        
        striation_features = {
            "num_striation_lines": 0,
            "dominant_directions": dominant_directions,
            "avg_line_length": 0.0,
            "striation_density": 0.0,
            "parallelism_score": 0.0
        }
        
        if lines is not None:
            striation_features["num_striation_lines"] = len(lines)
            
            line_lengths = []
            line_angles = []
            
            for line in lines:
                x1, y1, x2, y2 = line[0]
                length = np.sqrt((x2-x1)**2 + (y2-y1)**2)
                angle = np.arctan2(y2-y1, x2-x1) * 180 / np.pi
                
                line_lengths.append(length)
                line_angles.append(angle)
            
            if line_lengths:
                striation_features["avg_line_length"] = float(np.mean(line_lengths))
                striation_features["striation_density"] = len(lines) / (gray_img.shape[0] * gray_img.shape[1] / 10000)
                
                # Calcular score de paralelismo (varianza de ángulos)
                if len(line_angles) > 1:
                    angle_variance = np.var(line_angles)
                    striation_features["parallelism_score"] = float(1.0 / (1.0 + angle_variance/100))
        
        return striation_features
        
    except Exception as e:
        logger.warning(f"Error en detección de estriado: {str(e)}")
        return {
            "num_striation_lines": 0,
            "dominant_directions": [],
            "avg_line_length": 0.0,
            "striation_density": 0.0,
            "parallelism_score": 0.0
        }

def calculate_ballistic_features(image_data: bytes) -> dict:
    """Extrae características balísticas avanzadas incluyendo marcas de percutor y estriado"""
    try:
        nparr = np.frombuffer(image_data, np.uint8)
        img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
        if img is None:
            # Intentar con otros decodificadores si el formato no es estándar
            try:
                # Para formatos RAW, intentamos usar rawpy si está disponible
                import rawpy
                import io
                with rawpy.imread(io.BytesIO(image_data)) as raw:
                    img = raw.postprocess()
            except (ImportError, Exception) as e:
                logger.warning(f"No se pudo usar rawpy para decodificar: {str(e)}")
                raise FeatureExtractionError(f"Formato de imagen no soportado o requiere librería adicional: {str(e)}")
        
        # Verificar dimensiones de la imagen
        if img.shape[0] > 10000 or img.shape[1] > 10000:
            logger.warning(f"Imagen demasiado grande: {img.shape}")
            # Redimensionar para evitar problemas de memoria
            scale_factor = min(10000 / img.shape[0], 10000 / img.shape[1])
            new_width = int(img.shape[1] * scale_factor)
            new_height = int(img.shape[0] * scale_factor)
            img = cv2.resize(img, (new_width, new_height), interpolation=cv2.INTER_AREA)
            logger.info(f"Imagen redimensionada a: {img.shape}")
        
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
        
        # 1. Características tradicionales (momentos de Hu)
        blur = cv2.GaussianBlur(gray, (5, 5), 0)
        thresh = cv2.adaptiveThreshold(
            blur, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
            cv2.THRESH_BINARY_INV, 11, 2
        )
        
        contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        if not contours:
            logger.warning("No se encontraron contornos")
            hu_moments = [0.0] * 7
            contour_area = 0.0
            contour_len = 0.0
        else:
            largest_contour = max(contours, key=cv2.contourArea)
            mask = np.zeros_like(gray)
            cv2.drawContours(mask, [largest_contour], -1, 255, thickness=cv2.FILLED)
            
            masked_thresh = cv2.bitwise_and(thresh, thresh, mask=mask)
            moments = cv2.moments(masked_thresh)
            hu_moments = cv2.HuMoments(moments).flatten()
            
            # Calcular área y longitud del contorno
            contour_area = cv2.contourArea(largest_contour)
            contour_len = cv2.arcLength(largest_contour, True)
            
            hu_moments = np.where(
                np.abs(hu_moments) > 1e-9,
                -np.sign(hu_moments) * np.log10(np.abs(hu_moments)),
                0
            )
        
        # 2. Detectar marcas de percutor
        firing_pin_features = detect_firing_pin_marks(gray)
        
        # 3. Detectar patrones de estriado
        striation_features = detect_striation_patterns(gray)
        
        # 4. Características de textura mejoradas
        # Calcular LBP (Local Binary Patterns) para textura
        def calculate_lbp(image, radius=1, n_points=8):
            """Calcula Local Binary Pattern"""
            lbp = np.zeros_like(image, dtype=np.uint8)
            for i in range(radius, image.shape[0] - radius):
                for j in range(radius, image.shape[1] - radius):
                    center = image[i, j]
                    binary_value = 0
                    for k in range(n_points):
                        angle = 2 * np.pi * k / n_points
                        x = int(i + radius * np.cos(angle))
                        y = int(j + radius * np.sin(angle))
                        if x < image.shape[0] and y < image.shape[1]:
                            if image[x, y] >= center:
                                binary_value += (1 << k)
                    # Limitar el valor a 255 para uint8
                    lbp[i, j] = min(binary_value, 255)
            return lbp
        
        lbp = calculate_lbp(gray)
        lbp_hist, _ = np.histogram(lbp.flatten(), bins=256, range=(0, 256))
        lbp_uniformity = np.sum(lbp_hist**2) / (np.sum(lbp_hist)**2) if np.sum(lbp_hist) > 0 else 0
        
        # Combinar todas las características
        features = {
            "hu_moments": hu_moments.tolist(),
            "contour_area": float(contour_area),
            "contour_len": float(contour_len),
            "lbp_uniformity": float(lbp_uniformity),
            "firing_pin_marks": firing_pin_features,
            "striation_patterns": striation_features
        }
        
        return features
        
    except Exception as e:
        logger.exception("Error en cálculo de características balísticas")
        raise FeatureExtractionError(f"Error de procesamiento: {str(e)}")

def extract_features(image_path):
    """Extrae características de una imagen desde un archivo"""
    try:
        # Verificar si el archivo existe
        if not os.path.exists(image_path):
            raise FeatureExtractionError(f"Archivo no encontrado: {image_path}")
        
        # Verificar tamaño del archivo
        file_size = os.path.getsize(image_path)
        logger.info(f"Procesando archivo de {file_size/1024/1024:.2f} MB: {image_path}")
        
        # Intentar leer la imagen con OpenCV primero
        img = cv2.imread(image_path)
        
        # Si falla, intentar con otros métodos según la extensión
        if img is None:
            ext = os.path.splitext(image_path)[1].lower()
            
            if ext in [".tiff", ".tif"]:
                try:
                    from PIL import Image
                    with Image.open(image_path) as pil_img:
                        img = cv2.cvtColor(np.array(pil_img), cv2.COLOR_RGB2BGR)
                except Exception as e:
                    logger.warning(f"Error al leer TIFF con PIL: {str(e)}")
            
            elif ext in [".raw", ".cr2", ".nef", ".arw", ".dng"]:
                try:
                    import rawpy
                    with rawpy.imread(image_path) as raw:
                        img = raw.postprocess()
                        # Convertir de RGB a BGR para OpenCV
                        img = cv2.cvtColor(img, cv2.COLOR_RGB2BGR)
                except ImportError:
                    logger.error("rawpy no está instalado. Necesario para procesar archivos RAW")
                    raise FeatureExtractionError("Se requiere rawpy para procesar este formato de imagen")
                except Exception as e:
                    logger.error(f"Error al procesar archivo RAW: {str(e)}")
                    raise FeatureExtractionError(f"Error al procesar archivo RAW: {str(e)}")
        
        if img is None:
            raise FeatureExtractionError(f"No se pudo leer la imagen: {image_path}")
        
        # Registrar información sobre la imagen
        logger.info(f"Imagen cargada: {image_path}, dimensiones: {img.shape}")
        
        # Codificar la imagen para procesarla
        _, img_encoded = cv2.imencode('.png', img)
        image_data = img_encoded.tobytes()
        
        # Calcular características
        features = calculate_ballistic_features(image_data)
        return features
        
    except Exception as e:
        logger.exception(f"Error al procesar imagen: {image_path}")
        raise FeatureExtractionError(str(e))

@app.route('/extract', methods=['POST'])
def extract_features_endpoint():
    try:
        if 'image' not in request.files:
            return jsonify({"error": "No se proporcionó imagen"}), 400
            
        file = request.files['image']
        filename = file.filename
        content_type = file.content_type
        
        # Registrar información sobre el archivo recibido
        logger.info(f"Recibido archivo: {filename}, tipo: {content_type}")
        
        # Leer datos de la imagen
        image_data = file.read()
        file_size_mb = len(image_data) / (1024 * 1024)
        
        # Verificar tamaño del archivo
        if file_size_mb > 20:
            logger.warning(f"Archivo demasiado grande: {file_size_mb:.2f} MB")
            return jsonify({"error": f"El tamaño del archivo ({file_size_mb:.2f} MB) excede el límite de 20 MB"}), 413
        
        logger.info(f"Procesando imagen de {file_size_mb:.2f} MB")
        
        # Extraer características
        features = calculate_ballistic_features(image_data)
        
        # Añadir metadatos a la respuesta
        response = {
            "hu_moments": features["hu_moments"],
            "contour_area": features["contour_area"],
            "contour_len": features["contour_len"],
            "lbp_uniformity": features["lbp_uniformity"],
            "firing_pin_marks": features["firing_pin_marks"],
            "striation_patterns": features["striation_patterns"],
            "metadata": {
                "filename": filename,
                "content_type": content_type,
                "file_size_mb": file_size_mb
            },
            "status": "success"
        }
        
        return jsonify(response)
        
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

# Verificar dependencias opcionales al inicio
def check_optional_dependencies():
    missing_deps = []
    
    # Verificar rawpy para formatos RAW
    try:
        import rawpy
        logger.info("rawpy está disponible para procesamiento de archivos RAW")
    except ImportError:
        missing_deps.append("rawpy")
        logger.warning("rawpy no está instalado. El soporte para archivos RAW estará limitado")
    
    # Verificar PIL/Pillow para formatos adicionales
    try:
        from PIL import Image
        logger.info("PIL/Pillow está disponible para formatos de imagen adicionales")
    except ImportError:
        missing_deps.append("pillow")
        logger.warning("PIL/Pillow no está instalado. El soporte para algunos formatos estará limitado")
    
    return missing_deps

if __name__ == '__main__':
    # Verificar dependencias opcionales
    missing_deps = check_optional_dependencies()
    
    # Verificar si se está ejecutando como script de línea de comandos o como servidor Flask
    if len(sys.argv) > 1:
        # Modo línea de comandos para integración con Go
        parser = argparse.ArgumentParser(description='Extractor de características de imágenes balísticas')
        parser.add_argument('image_path', type=str, help='Ruta a la imagen para analizar')
        args = parser.parse_args()
        
        try:
            # Extraer características
            features = extract_features(args.image_path)
            
            # Obtener información del archivo
            file_size = os.path.getsize(args.image_path) / (1024 * 1024)
            
            # Convertir a formato compatible con PythonResponse en Go
            response = {
                "hu_moments": features["hu_moments"],
                "contour_area": features["contour_area"],
                "contour_len": features["contour_len"],
                "lbp_uniformity": features["lbp_uniformity"],
                "firing_pin_marks": features["firing_pin_marks"],
                "striation_patterns": features["striation_patterns"],
                "metadata": {
                    "filename": os.path.basename(args.image_path),
                    "file_size_mb": file_size
                }
            }
            
            # Imprimir como JSON para que Go pueda parsearlo
            print(json.dumps(response))
            sys.exit(0)
        except Exception as e:
            # Imprimir error en formato JSON para que Go pueda parsearlo
            print(json.dumps({"error": str(e)}))
            sys.exit(1)
    else:
        # Modo servidor Flask
        app.run(host='0.0.0.0', port=5000)
    