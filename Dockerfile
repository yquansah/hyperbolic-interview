FROM golang:1.21.6-alpine

WORKDIR /app

COPY go.* .
COPY main.go .

RUN go build -o /hyperbolic

CMD ["/hyperbolic"]
