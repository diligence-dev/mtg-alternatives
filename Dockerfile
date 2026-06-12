FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go install golang.org/dl/go1.26.4@latest
RUN go1.26.4 download
RUN go1.26.4 mod download
COPY . .
RUN CGO_ENABLED=1 go1.26.4 build -o mtg-alternatives .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/mtg-alternatives .
EXPOSE 8080
CMD ["./mtg-alternatives"]
