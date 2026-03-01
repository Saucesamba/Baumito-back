FROM golang:1.26.0-alpine3.23 AS builder

ARG SERVICE_PATH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service ./${SERVICE_PATH}/main.go

FROM golang:1.26.0-alpine3.23

WORKDIR /root/

COPY --from=builder /app/service .
COPY --from=builder /app/configs ./configs

CMD ["./service"]
