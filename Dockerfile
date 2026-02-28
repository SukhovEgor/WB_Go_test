FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /main ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /main .
COPY frontend ./frontend

EXPOSE 8080
CMD ["./main"]