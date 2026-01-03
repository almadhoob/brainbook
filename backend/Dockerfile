FROM golang:1.25 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o brainbook-api .

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=build /app/brainbook-api /app/brainbook-api

ENV HTTP_PORT=8080
EXPOSE 8080

CMD ["/app/brainbook-api"]
