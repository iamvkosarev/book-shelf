# ==== Build ====
FROM golang:1.24-bullseye AS builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o service ./cmd/main.go

# ==== Runtime ====
FROM scratch AS service

WORKDIR /

COPY --from=builder /app/service /service

EXPOSE ${HTTP_PORT}
ENTRYPOINT ["/service"]