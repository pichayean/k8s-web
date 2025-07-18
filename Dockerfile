FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go kube.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/server .
COPY templates ./templates
COPY static ./static

EXPOSE 4000

CMD ["./server"]
