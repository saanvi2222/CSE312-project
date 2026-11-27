FROM golang:1.27

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY app/ ./app/

WORKDIR /app/app

RUN go build -o server .

EXPOSE 8080

CMD ["./server"]
