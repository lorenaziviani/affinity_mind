FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o backend main.go

FROM golang:1.20-alpine AS dev
WORKDIR /app
COPY . .
RUN go mod tidy
CMD ["tail", "-f", "/dev/null"]

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/backend .
EXPOSE 8080
CMD ["./backend"]