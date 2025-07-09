# 🧱 Builder stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ✅ build แบบ static
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go kube.go

# 🐧 Minimal runtime
FROM scratch

WORKDIR /app

COPY --from=builder /app/server .
COPY templates ./templates

EXPOSE 4000

CMD ["./server"]
