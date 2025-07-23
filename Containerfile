FROM golang:1.24-alpine
LABEL authors="Zymantas Maumevicius"

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o secretary -ldflags="-s -w" ./cmd/secretary

ENTRYPOINT ["./secretary"]
