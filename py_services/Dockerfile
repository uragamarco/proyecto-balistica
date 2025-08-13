ARG PYTHON_VERSION=3.9
FROM python:${PYTHON_VERSION}-slim-buster

# Configuración del entorno
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=off \
    OPENCV_THREADS=4

# Instalar dependencias del sistema
RUN apt-get update && apt-get install -y \
    build-essential \
    libgl1-mesa-glx \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    && rm -rf /var/lib/apt/lists/*

# Configurar directorio de trabajo
WORKDIR /app

# Copiar e instalar dependencias primero para aprovechar caché
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar código de la aplicación
COPY . .

# Crear directorio para datos temporales
RUN mkdir -p /tmp/balistica && chmod 777 /tmp/balistica

# Exponer puerto del servicio
EXPOSE 5000

# Comando de inicio con soporte para health checks
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:5000", "feature_extractor:app"]
