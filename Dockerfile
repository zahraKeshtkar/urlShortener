FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /urlshortener

FROM alpine:latest

WORKDIR /app

COPY --from=builder /urlshortener app/config.yaml  ./

EXPOSE 8080

CMD ./urlshortener setupdb ; ./urlshortener server
