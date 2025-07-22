FROM golang:1.24-alpine
LABEL authors="Zymantas Maumevicius"

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o secretary -ldflags="-s -w" ./...

ENTRYPOINT ["./secretary"]
