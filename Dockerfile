FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/app ./cmd/server
RUN wget -O GeoLite2-Country.mmdb "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"

FROM builder AS development
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM alpine:3.23.3 AS production
WORKDIR /app
COPY --from=builder /app/bin/app .
COPY --from=builder /app/GeoLite2-Country.mmdb .
EXPOSE 8080
CMD ["./app"]