FROM golang:1.23-alpine

RUN apk add --no-cache git bash curl build-base
RUN go install github.com/air-verse/air@v1.61.7
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
