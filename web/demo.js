// Estado global de la aplicación
let selectedFile = null;
let currentAnalysisResult = null;
let analysisHistory = [];

// Inicialización cuando se carga la página
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

function initializeApp() {
    setupNavigation();
    setupFileUpload();
    setupEventListeners();
    loadAnalysisHistory();
}

// === NAVEGACIÓN ===
function setupNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    
    navItems.forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            const targetSection = this.getAttribute('data-section');
            showSection(targetSection);
            
            // Actualizar navegación activa
            navItems.forEach(nav => nav.classList.remove('active'));
            this.classList.add('active');
        });
    });
}

function showSection(sectionName) {
    const sections = document.querySelectorAll('.content-section');
    sections.forEach(section => section.classList.remove('active'));
    
    const targetSection = document.getElementById(sectionName + '-section');
    if (targetSection) {
        targetSection.classList.add('active');
    }
}

// === CARGA DE ARCHIVOS ===
function setupFileUpload() {
    const uploadArea = document.getElementById('uploadArea');
    const fileInput = document.getElementById('fileInput');
    const uploadBtn = document.getElementById('uploadBtn');
    
    // Click en área de carga
    uploadArea.addEventListener('click', () => fileInput.click());
    uploadBtn.addEventListener('click', (e) => {
        e.stopPropagation(); // Evitar propagación del evento
        fileInput.click();
    });
    
    // Drag and drop
    uploadArea.addEventListener('dragover', handleDragOver);
    uploadArea.addEventListener('dragleave', handleDragLeave);
    uploadArea.addEventListener('drop', handleDrop);
    
    // Selección de archivo
    fileInput.addEventListener('change', handleFileSelect);
}

function handleDragOver(e) {
    e.preventDefault();
    e.currentTarget.classList.add('dragover');
}

function handleDragLeave(e) {
    e.preventDefault();
    e.currentTarget.classList.remove('dragover');
}

function handleDrop(e) {
    e.preventDefault();
    e.currentTarget.classList.remove('dragover');
    
    const files = e.dataTransfer.files;
    if (files.length > 0) {
        handleFile(files[0]);
    }
}

function handleFileSelect(event) {
    const file = event.target.files[0];
    if (file) {
        handleFile(file);
    }
}

function handleFile(file) {
    // Validar tipo de archivo
    if (!file.type.startsWith('image/')) {
        showError('Por favor selecciona un archivo de imagen válido.');
        return;
    }
    
    // Validar tamaño (máximo 20MB)
    if (file.size > 20 * 1024 * 1024) {
        showError('El archivo es demasiado grande. Máximo 20MB.');
        return;
    }
    
    selectedFile = file;
    showImagePreview(file);
}

function showImagePreview(file) {
    const uploadArea = document.getElementById('uploadArea');
    const imagePreview = document.getElementById('imagePreview');
    const previewImage = document.getElementById('previewImage');
    const fileName = document.getElementById('fileName');
    const fileSize = document.getElementById('fileSize');
    const imageDimensions = document.getElementById('imageDimensions');
    
    // Ocultar área de carga y mostrar preview
    uploadArea.style.display = 'none';
    imagePreview.style.display = 'block';
    
    // Mostrar imagen
    const reader = new FileReader();
    reader.onload = function(e) {
        previewImage.src = e.target.result;
        
        // Obtener dimensiones de la imagen
        previewImage.onload = function() {
            imageDimensions.textContent = `${this.naturalWidth} x ${this.naturalHeight}px`;
        };
    };
    reader.readAsDataURL(file);
    
    // Mostrar información del archivo
    fileName.textContent = file.name;
    fileSize.textContent = formatFileSize(file.size);
}

function changeImage() {
    document.getElementById('uploadArea').style.display = 'block';
    document.getElementById('imagePreview').style.display = 'none';
    document.getElementById('loadingContainer').style.display = 'none';
    document.getElementById('resultsContainer').style.display = 'none';
    selectedFile = null;
    currentAnalysisResult = null;
}

// === ANÁLISIS ===
function analyzeImage() {
    if (!selectedFile) {
        showError('Por favor selecciona una imagen primero.');
        return;
    }
    
    showLoading();
    
    const formData = new FormData();
    formData.append('image', selectedFile);
    
    // Crear AbortController para timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 60000); // 60 segundos timeout
    
    fetch('/api/process', {
        method: 'POST',
        body: formData,
        signal: controller.signal
    })
    .then(response => {
        clearTimeout(timeoutId);
        if (!response.ok) {
            throw new Error(`Error ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        currentAnalysisResult = data;
        hideLoading();
        displayResults(data);
        saveToHistory(data);
    })
    .catch(error => {
        clearTimeout(timeoutId);
        console.error('Error:', error);
        hideLoading();
        if (error.name === 'AbortError') {
            showError('El análisis tardó demasiado tiempo. Por favor, intenta con una imagen más pequeña o verifica tu conexión.');
        } else {
            showError('Error al analizar la imagen: ' + error.message);
        }
    });
}

function showLoading() {
    document.getElementById('imagePreview').style.display = 'none';
    document.getElementById('loadingContainer').style.display = 'block';
    document.getElementById('resultsContainer').style.display = 'none';
}

function hideLoading() {
    document.getElementById('loadingContainer').style.display = 'none';
}

// === RESULTADOS ===
function displayResults(data) {
    document.getElementById('resultsContainer').style.display = 'block';
    
    // Clasificación automática
    displayClassification(data.classification);
    
    // Características detectadas
    displayFeatures(data.features);
    
    // Detalles técnicos
    displayTechnicalDetails(data);
}

function displayClassification(classification) {
    if (!classification) return;
    
    const weaponType = document.getElementById('weaponType');
    const weaponConfidence = document.getElementById('weaponConfidence');
    const weaponConfidenceBar = document.getElementById('weaponConfidenceBar');
    const weaponConfidenceText = document.getElementById('weaponConfidenceText');
    
    const caliber = document.getElementById('caliber');
    const caliberConfidence = document.getElementById('caliberConfidence');
    const caliberConfidenceBar = document.getElementById('caliberConfidenceBar');
    const caliberConfidenceText = document.getElementById('caliberConfidenceText');
    
    // Tipo de arma
    weaponType.textContent = classification.weapon_type || 'No determinado';
    const weaponConf = (classification.confidence || 0) * 100;
    weaponConfidence.textContent = weaponConf.toFixed(1) + '%';
    weaponConfidenceBar.style.width = weaponConf + '%';
    weaponConfidenceBar.className = 'confidence-fill ' + getConfidenceClass(weaponConf);
    weaponConfidenceText.textContent = getConfidenceText(weaponConf);
    weaponConfidenceText.className = 'confidence-text ' + getConfidenceClass(weaponConf);
    
    // Calibre
    caliber.textContent = classification.caliber || 'No determinado';
    const caliberConf = (classification.caliber_confidence || classification.confidence || 0) * 100;
    caliberConfidence.textContent = caliberConf.toFixed(1) + '%';
    caliberConfidenceBar.style.width = caliberConf + '%';
    caliberConfidenceBar.className = 'confidence-fill ' + getConfidenceClass(caliberConf);
    caliberConfidenceText.textContent = getConfidenceText(caliberConf);
    caliberConfidenceText.className = 'confidence-text ' + getConfidenceClass(caliberConf);
}

function displayFeatures(features) {
    if (!features) return;
    
    const featuresGrid = document.getElementById('featuresGrid');
    featuresGrid.innerHTML = '';
    
    // Mapear los datos reales del backend a las características mostradas
    const featureTypes = [
        { 
            key: 'firing_pin_count', 
            name: 'Marcas de Percutor', 
            icon: 'fas fa-bullseye',
            getValue: (f) => f.firing_pin_count || 0,
            getDetail: (f) => f.firing_pin_count > 0 ? `${f.firing_pin_count} marcas detectadas (radio promedio: ${(f.firing_pin_avg_radius || 0).toFixed(1)}px)` : 'No detectado'
        },
        { 
            key: 'contour_area', 
            name: 'Marcas de Recámara', 
            icon: 'fas fa-circle',
            getValue: (f) => f.contour_area || 0,
            getDetail: (f) => f.contour_area > 0 ? `Área de contorno: ${f.contour_area.toLocaleString()} píxeles` : 'No detectado'
        },
        { 
            key: 'glcm_contrast', 
            name: 'Marcas de Extractor', 
            icon: 'fas fa-grip-lines',
            getValue: (f) => f.glcm_contrast || 0,
            getDetail: (f) => f.glcm_contrast > 0 ? `Contraste GLCM: ${f.glcm_contrast.toFixed(2)}` : 'No detectado'
        },
        { 
            key: 'circularity', 
            name: 'Marcas de Eyector', 
            icon: 'fas fa-arrow-right',
            getValue: (f) => f.circularity || 0,
            getDetail: (f) => f.circularity > 0 ? `Circularidad: ${(f.circularity * 100).toFixed(2)}%` : 'No detectado'
        },
        { 
            key: 'lbp_uniformity', 
            name: 'Marcas de Cámara', 
            icon: 'fas fa-ring',
            getValue: (f) => f.lbp_uniformity || 0,
            getDetail: (f) => f.lbp_uniformity > 0 ? `Uniformidad LBP: ${(f.lbp_uniformity * 100).toFixed(1)}%` : 'No detectado'
        },
        { 
            key: 'striation_count', 
            name: 'Patrones de Estriado', 
            icon: 'fas fa-wave-square',
            getValue: (f) => f.striation_count || 0,
            getDetail: (f) => f.striation_count > 0 ? `${f.striation_count} patrones detectados` : 'No detectado'
        }
    ];
    
    featureTypes.forEach(featureType => {
        const value = featureType.getValue(features);
        const detected = value > 0;
        const detail = featureType.getDetail(features);
        
        const featureItem = document.createElement('div');
        featureItem.className = 'feature-item';
        featureItem.innerHTML = `
            <div class="feature-icon ${detected ? 'detected' : 'not-detected'}">
                <i class="${featureType.icon}"></i>
            </div>
            <div class="feature-info">
                <h5>${featureType.name}</h5>
                <p>${detail}</p>
            </div>
        `;
        
        featuresGrid.appendChild(featureItem);
    });
}

function displayTechnicalDetails(data) {
    const detailsContent = document.getElementById('detailsContent');
    detailsContent.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
}

function toggleTechnicalDetails() {
    const detailsContent = document.getElementById('detailsContent');
    const toggleIcon = document.getElementById('toggleIcon');
    
    if (detailsContent.style.display === 'none' || !detailsContent.style.display) {
        detailsContent.style.display = 'block';
        toggleIcon.classList.add('rotated');
    } else {
        detailsContent.style.display = 'none';
        toggleIcon.classList.remove('rotated');
    }
}

// === UTILIDADES ===
function getConfidenceClass(confidence) {
    if (confidence >= 80) return 'high';
    if (confidence >= 50) return 'medium';
    return 'low';
}

function getConfidenceText(confidence) {
    if (confidence >= 80) return 'Alta confianza';
    if (confidence >= 50) return 'Confianza media';
    return 'Baja confianza';
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function showError(message) {
    alert(message); // Temporal - se puede mejorar con un modal
}

function setupEventListeners() {
    // Botón de análisis
    const analyzeBtn = document.getElementById('analyzeBtn');
    if (analyzeBtn) {
        analyzeBtn.addEventListener('click', analyzeImage);
    }
    
    // Botón de cambiar imagen
    const changeBtn = document.getElementById('changeImageBtn');
    if (changeBtn) {
        changeBtn.addEventListener('click', changeImage);
    }
    
    // Toggle de detalles técnicos
    const detailsHeader = document.getElementById('detailsHeader');
    if (detailsHeader) {
        detailsHeader.addEventListener('click', toggleTechnicalDetails);
    }
}

// === HISTORIAL ===
function saveToHistory(result) {
    const historyItem = {
        id: Date.now(),
        timestamp: new Date().toISOString(),
        fileName: selectedFile ? selectedFile.name : 'Imagen sin nombre',
        classification: result.classification,
        features: result.features
    };
    
    analysisHistory.unshift(historyItem);
    
    // Mantener solo los últimos 50 análisis
    if (analysisHistory.length > 50) {
        analysisHistory = analysisHistory.slice(0, 50);
    }
    
    localStorage.setItem('ballistic_analysis_history', JSON.stringify(analysisHistory));
}

function loadAnalysisHistory() {
    const saved = localStorage.getItem('ballistic_analysis_history');
    if (saved) {
        try {
            analysisHistory = JSON.parse(saved);
        } catch (e) {
            console.error('Error loading history:', e);
            analysisHistory = [];
        }
    }
}

// === FUNCIONES DE EXPORTACIÓN ===
function downloadResults() {
    if (!currentAnalysisResult) {
        showError('No hay resultados para descargar.');
        return;
    }
    
    const dataStr = JSON.stringify(currentAnalysisResult, null, 2);
    const dataBlob = new Blob([dataStr], {type: 'application/json'});
    
    const link = document.createElement('a');
    link.href = URL.createObjectURL(dataBlob);
    link.download = `analisis_balistico_${new Date().toISOString().split('T')[0]}.json`;
    link.click();
}

function shareResults() {
    if (!currentAnalysisResult) {
        showError('No hay resultados para compartir');
        return;
    }
    
    const shareData = {
        title: 'Análisis Balístico - Proyecto Balística',
        text: `Análisis completado: ${currentAnalysisResult.classification?.weapon_type || 'Tipo no determinado'}`,
        url: window.location.href
    };
    
    if (navigator.share) {
        navigator.share(shareData).catch(err => {
            console.log('Error al compartir:', err);
            fallbackShare();
        });
    } else {
        fallbackShare();
    }
    
    function fallbackShare() {
        navigator.clipboard.writeText(window.location.href).then(() => {
            alert('Enlace copiado al portapapeles');
        }).catch(() => {
            alert('No se pudo copiar el enlace');
        });
    }
}

// === FUNCIONALIDAD DE COMPARACIÓN ===
let comparisonFiles = {
    sample1: null,
    sample2: null
};

function setupComparisonUpload() {
    setupSampleUpload('sample1');
    setupSampleUpload('sample2');
    
    const compareBtn = document.getElementById('compareBtn');
    if (compareBtn) {
        compareBtn.addEventListener('click', performComparison);
    }
}

function setupSampleUpload(sampleId) {
    const uploadBox = document.getElementById(`${sampleId}Upload`);
    const fileInput = document.getElementById(`${sampleId}Input`);
    const previewArea = document.getElementById(`${sampleId}Preview`);
    const changeBtn = document.getElementById(`${sampleId}Change`);
    
    if (!uploadBox || !fileInput) return;
    
    // Click para seleccionar archivo
    uploadBox.addEventListener('click', () => {
        if (!comparisonFiles[sampleId]) {
            fileInput.click();
        }
    });
    
    // Drag and drop
    uploadBox.addEventListener('dragover', handleComparisonDragOver);
    uploadBox.addEventListener('dragleave', handleComparisonDragLeave);
    uploadBox.addEventListener('drop', (e) => handleComparisonDrop(e, sampleId));
    
    // Selección de archivo
    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            handleComparisonFile(e.target.files[0], sampleId);
        }
    });
    
    // Botón cambiar
    if (changeBtn) {
        changeBtn.addEventListener('click', () => {
            comparisonFiles[sampleId] = null;
            fileInput.value = '';
            showComparisonUploadBox(sampleId);
            updateCompareButton();
        });
    }
}

function handleComparisonDragOver(e) {
    e.preventDefault();
    e.currentTarget.classList.add('dragover');
}

function handleComparisonDragLeave(e) {
    e.preventDefault();
    e.currentTarget.classList.remove('dragover');
}

function handleComparisonDrop(e, sampleId) {
    e.preventDefault();
    e.currentTarget.classList.remove('dragover');
    
    const files = e.dataTransfer.files;
    if (files.length > 0) {
        handleComparisonFile(files[0], sampleId);
    }
}

function handleComparisonFile(file, sampleId) {
    // Validar tipo de archivo
    const allowedTypes = ['image/jpeg', 'image/png', 'image/tiff'];
    if (!allowedTypes.includes(file.type)) {
        showError('Por favor selecciona un archivo de imagen válido (JPEG, PNG, TIFF)');
        return;
    }
    
    // Validar tamaño (20MB)
    const maxSize = 20 * 1024 * 1024;
    if (file.size > maxSize) {
        showError('El archivo es demasiado grande. Máximo 20MB.');
        return;
    }
    
    comparisonFiles[sampleId] = file;
    showComparisonPreview(file, sampleId);
    updateCompareButton();
}

function showComparisonPreview(file, sampleId) {
    const uploadBox = document.getElementById(`${sampleId}Upload`);
    const previewArea = document.getElementById(`${sampleId}Preview`);
    
    if (!uploadBox || !previewArea) return;
    
    const reader = new FileReader();
    reader.onload = function(e) {
        uploadBox.style.display = 'none';
        previewArea.style.display = 'block';
        
        const img = previewArea.querySelector('img');
        const fileName = previewArea.querySelector('.file-name');
        const fileSize = previewArea.querySelector('.file-size');
        
        if (img) img.src = e.target.result;
        if (fileName) fileName.textContent = file.name;
        if (fileSize) fileSize.textContent = formatFileSize(file.size);
    };
    reader.readAsDataURL(file);
}

function showComparisonUploadBox(sampleId) {
    const uploadBox = document.getElementById(`${sampleId}Upload`);
    const previewArea = document.getElementById(`${sampleId}Preview`);
    
    if (uploadBox) uploadBox.style.display = 'block';
    if (previewArea) previewArea.style.display = 'none';
}

function updateCompareButton() {
    const compareBtn = document.getElementById('compareBtn');
    if (!compareBtn) return;
    
    const canCompare = comparisonFiles.sample1 && comparisonFiles.sample2;
    compareBtn.disabled = !canCompare;
    compareBtn.textContent = canCompare ? 'Comparar Muestras' : 'Selecciona ambas muestras';
}

function performComparison() {
    if (!comparisonFiles.sample1 || !comparisonFiles.sample2) {
        showError('Por favor selecciona ambas muestras para comparar');
        return;
    }
    
    showComparisonLoading();
    
    const formData = new FormData();
    formData.append('sample1', comparisonFiles.sample1);
    formData.append('sample2', comparisonFiles.sample2);
    
    fetch('/api/compare', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Error ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        hideComparisonLoading();
        displayComparisonResults(data);
    })
    .catch(error => {
        hideComparisonLoading();
        console.error('Error en la comparación:', error);
        showError('Error al procesar la comparación: ' + error.message);
    });
}

function showComparisonLoading() {
    const loadingDiv = document.getElementById('comparisonLoading');
    const resultsDiv = document.getElementById('comparisonResults');
    
    if (loadingDiv) loadingDiv.style.display = 'block';
    if (resultsDiv) resultsDiv.style.display = 'none';
}

function hideComparisonLoading() {
    const loadingDiv = document.getElementById('comparisonLoading');
    if (loadingDiv) loadingDiv.style.display = 'none';
}

function displayComparisonResults(data) {
    const resultsDiv = document.getElementById('comparisonResults');
    if (!resultsDiv) return;
    
    resultsDiv.style.display = 'block';
    
    // Actualizar puntuación de similitud
    updateSimilarityScore(data.similarity_score || 0);
    
    // Actualizar comparación de características
    updateFeatureComparison(data.feature_comparison || {});
    
    // Actualizar análisis estadístico
    updateStatisticalAnalysis(data.statistical_analysis || {});
    
    // Actualizar conclusión
    updateComparisonConclusion(data.conclusion || {});
}

function updateSimilarityScore(score) {
    const scoreValue = document.getElementById('similarityValue');
    const scoreCircle = document.querySelector('.score-circle');
    const confidenceText = document.getElementById('confidenceLevel');
    
    if (scoreValue) {
        scoreValue.textContent = `${score.toFixed(1)}%`;
    }
    
    if (scoreCircle) {
        const percentage = score / 100;
        const degrees = percentage * 360;
        scoreCircle.style.background = `conic-gradient(var(--primary-color) ${degrees}deg, #e9ecef ${degrees}deg)`;
    }
    
    if (confidenceText) {
        let confidence, confidenceClass;
        if (score >= 80) {
            confidence = 'Alta confianza';
            confidenceClass = 'high';
        } else if (score >= 50) {
            confidence = 'Confianza media';
            confidenceClass = 'medium';
        } else {
            confidence = 'Baja confianza';
            confidenceClass = 'low';
        }
        
        confidenceText.textContent = confidence;
        confidenceText.className = `confidence-text ${confidenceClass}`;
    }
}

function updateFeatureComparison(features) {
    const container = document.getElementById('featureComparisonList');
    if (!container) return;
    
    container.innerHTML = '';
    
    const featureNames = {
        firing_pin: 'Marcas de Percutor',
        chamber: 'Marcas de Recámara',
        extractor: 'Marcas de Extractor',
        ejector: 'Marcas de Eyector',
        breech_face: 'Marcas de Cámara',
        rifling: 'Patrones de Estriado'
    };
    
    Object.entries(features).forEach(([key, value]) => {
        const item = document.createElement('div');
        item.className = 'feature-comparison-item';
        
        const similarity = value.similarity || 0;
        let similarityClass;
        if (similarity >= 80) similarityClass = 'high';
        else if (similarity >= 50) similarityClass = 'medium';
        else similarityClass = 'low';
        
        item.innerHTML = `
            <span class="feature-name">${featureNames[key] || key}</span>
            <span class="feature-similarity ${similarityClass}">${similarity.toFixed(1)}%</span>
        `;
        
        container.appendChild(item);
    });
}

function updateStatisticalAnalysis(analysis) {
    const container = document.getElementById('statisticalMetrics');
    if (!container) return;
    
    container.innerHTML = '';
    
    const metrics = {
        correlation: 'Correlación',
        variance: 'Varianza',
        standard_deviation: 'Desviación Estándar',
        mean_difference: 'Diferencia Media',
        confidence_interval: 'Intervalo de Confianza'
    };
    
    Object.entries(analysis).forEach(([key, value]) => {
        const item = document.createElement('div');
        item.className = 'statistical-metric';
        
        let displayValue = value;
        if (typeof value === 'number') {
            displayValue = value.toFixed(3);
        }
        
        item.innerHTML = `
            <span class="metric-name">${metrics[key] || key}</span>
            <span class="metric-value">${displayValue}</span>
        `;
        
        container.appendChild(item);
    });
}

function updateComparisonConclusion(conclusion) {
    const container = document.getElementById('conclusionText');
    if (!container) return;
    
    const text = conclusion.text || 'Análisis completado. Los resultados muestran las similitudes y diferencias entre las muestras analizadas.';
    const highlight = conclusion.highlight || '';
    
    container.innerHTML = `
        <p class="conclusion-text">${text}</p>
        ${highlight ? `<div class="conclusion-highlight">${highlight}</div>` : ''}
    `;
}

// Inicializar comparación cuando se muestra la sección
document.addEventListener('DOMContentLoaded', function() {
    setupComparisonUpload();
});
