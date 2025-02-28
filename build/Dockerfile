FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

RUN go install golang.org/x/vuln/cmd/govulncheck@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go vet ./...

RUN govulncheck ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o pgc ./cmd/pgc

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/pgc .
COPY migrations/ ./migrations/

CMD ["./pgc"]
