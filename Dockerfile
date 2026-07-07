FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myapp .

FROM alpine:3.19

WORKDIR /root/

COPY --from=builder /app/myapp .

EXPOSE 8080

CMD ["./myapp"]
