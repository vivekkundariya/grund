FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/main.go
EXPOSE 8081
CMD ["./main"]
