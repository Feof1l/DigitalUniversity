FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o digitalUniversity .

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/digitalUniversity .

CMD ["./digitalUniversity"]
