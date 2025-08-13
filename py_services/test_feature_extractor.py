import os
import pytest
import numpy as np
import cv2
from feature_extractor import calculate_hu_moments, extract_features, FeatureExtractionError

# Crear imagen de prueba en memoria
def create_test_image(width=500, height=500, circles=3):
    img = np.zeros((height, width, 3), dtype=np.uint8)
    for i in range(1, circles + 1):
        center = (width//2, height//2)
        radius = (width//4) // i
        cv2.circle(img, center, radius, (255, 255, 255), -1)
    _, img_encoded = cv2.imencode('.png', img)
    return img_encoded.tobytes()

@pytest.fixture
def test_image_bytes():
    return create_test_image()

@pytest.fixture
def test_image_file(tmp_path):
    img_bytes = create_test_image()
    img_path = tmp_path / "test_image.png"
    with open(img_path, 'wb') as f:
        f.write(img_bytes)
    return str(img_path)

def test_calculate_hu_moments_valid_image(test_image_bytes):
    hu_moments = calculate_hu_moments(test_image_bytes)
    assert len(hu_moments) == 7
    # Verificar que no sean todos cero
    assert not all(hu == 0.0 for hu in hu_moments)
    # Verificar normalización
    assert all(not math.isnan(hu) for hu in hu_moments)

def test_calculate_hu_moments_invalid_image():
    with pytest.raises(FeatureExtractionError):
        calculate_hu_moments(b'invalid_image_data')

def test_calculate_hu_moments_empty_image():
    empty_img = np.zeros((100, 100, 3), dtype=np.uint8)
    _, img_encoded = cv2.imencode('.png', empty_img)
    with pytest.raises(FeatureExtractionError):
        calculate_hu_moments(img_encoded.tobytes())

def test_extract_features_valid_image(test_image_file):
    # Configuración mínima
    config = {
        'imaging': {
            'min_quality': 90
        },
        'features': {
            'hu_moments_scaling': 'logarithmic'
        }
    }
    
    features = extract_features(test_image_file, config)
    assert isinstance(features, dict)
    assert 'hu_moment_1' in features
    assert 'contour_area' in features
    assert features['contour_area'] > 0

def test_extract_features_nonexistent_file():
    with pytest.raises(FeatureExtractionError):
        extract_features("nonexistent_file.png", {})

def test_extract_features_invalid_file(tmp_path):
    invalid_file = tmp_path / "invalid.txt"
    with open(invalid_file, 'w') as f:
        f.write("This is not an image")
    
    with pytest.raises(FeatureExtractionError):
        extract_features(str(invalid_file), {})

def test_health_check(client):
    response = client.get('/health')
    assert response.status_code == 200
    assert response.json == {"status": "healthy", "service": "feature-extractor"}
    