FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka ./scripts/kafka
RUN chmod +x /kafka


FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /kafka /kafka
WORKDIR /app
EXPOSE 8082
CMD ["/kafka"]