# Build stage
FROM golang:1.22 as builder

WORKDIR /app
COPY . .

RUN apt-get update && apt-get install -y libopencv-dev
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /proyecto-balistica cmd/main.go

# Runtime stage
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y libopencv-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /proyecto-balistica .
COPY configs/default.yml ./configs/
COPY .env .

EXPOSE 8080

CMD ["./proyecto-balistica"]