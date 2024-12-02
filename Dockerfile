FROM golang:alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/ubiquitous-octo-fiesta/main.go

FROM alpine:latest as final
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
EXPOSE 8082
RUN /bin/sh
CMD ["./main"]