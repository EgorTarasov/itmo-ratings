FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o rating-scrapper ./cmd/rating-scrapper
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o service ./cmd/service

FROM alpine:3.20
# Accept commit SHA at build time and expose as runtime env var (can be overridden when running the container)
ARG COMMIT_SHA=unknown
ENV COMMIT_ID=$COMMIT_SHA
LABEL org.opencontainers.image.revision=$COMMIT_SHA
RUN apk --no-cache add ca-certificates && adduser -D -g "" appuser

WORKDIR /root/

COPY --from=builder /app/rating-scrapper /usr/local/bin/
COPY --from=builder /app/service /usr/local/bin/

USER appuser
ENTRYPOINT ["service"]