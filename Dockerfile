FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o pubbet ./cmd/

FROM alpine:latest AS runner

WORKDIR /app

COPY --from=builder /app/pubbet ./pubbet

RUN addgroup -S pubbet && adduser -S pubbet -G pubbet
USER pubbet

CMD ["./pubbet"]