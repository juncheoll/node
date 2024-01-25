FROM golang:alpine

COPY . /app

WORKDIR /app

RUN go mod download
RUN go build main.go 

CMD ["./main"]