server:
  address: ":8080"
  timeout:
    read: 10s
    write: 30s
    idle: 60s

database:
  chroma:
    url: "http://localhost:8000"
    collection: "balistica"

security:
  jwt_secret: "cambiar-por-una-clave-segura"
  api_keys:
    - "clave-desarrollo-1"
    - "clave-desarrollo-2"

limits:
  upload:
    max_size_mb: 10
    requests_per_minute: 5
  global:
    requests_per_second: 20