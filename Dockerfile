FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-pickup-point-api ./cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /go-pickup-point-api /go-pickup-point-api

COPY config/config.yaml /config/config.yaml

COPY .env /.env


ENTRYPOINT ["/go-pickup-point-api"]
CMD ["/config/config.yaml"]