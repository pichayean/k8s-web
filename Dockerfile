FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server main.go

FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/server .
COPY templates ./templates

EXPOSE 4000

CMD ["./server"]
