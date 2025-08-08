document.getElementById('analyzeBtn').addEventListener('click', async () => {
    const fileInput = document.getElementById('imageInput');
    const file = fileInput.files[0];
    
    if (!file) {
        alert('Por favor selecciona una imagen');
        return;
    }

    const formData = new FormData();
    formData.append('image', file);

    try {
        const response = await fetch('http://localhost:8080/api/process', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            throw new Error('Error en el servidor');
        }

        const data = await response.json();
        
        // Mostrar características
        document.getElementById('features').innerHTML = `
            <p><strong>Similitud:</strong> ${data.similarity?.toFixed(2) || 'N/A'}</p>
            <p><strong>Colores dominantes:</strong> ${data.chroma_data?.dominant_colors?.length || 0} detectados</p>
        `;

        // Mostrar imagen procesada (si está disponible en la respuesta)
        if (data.processed_image_url) {
            document.getElementById('processedImage').src = data.processed_image_url;
        }

    } catch (error) {
        console.error('Error:', error);
        alert('Error al procesar la imagen: ' + error.message);
    }
});
