
FROM golang:latest

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o srv cmd/server/main.go 

EXPOSE 5001

CMD ["/app/srv"]
