# 🌱 Builder stage
FROM golang:1.22 AS builder

WORKDIR /app

# คัดลอก dependency files ก่อน เพื่อใช้ layer cache
COPY go.mod go.sum ./
RUN go mod download

# คัดลอก source code
COPY . .

# 🔧 Compile go app
RUN go build -o server main.go kube.go

# 🐧 Runtime stage
FROM debian:bullseye-slim

WORKDIR /app

# 🔐 สำหรับ certs ที่ client-go อาจใช้ (เช่นเมื่อ K8s cluster ใช้ TLS)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# 🔁 คัดลอก binary และ templates
COPY --from=builder /app/server .
COPY templates ./templates

EXPOSE 4000

CMD ["./server"]
