FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o rating-scrapper ./cmd/rating-scrapper

FROM alpine:latest
RUN apk --no-cache add ca-certificates dcron

WORKDIR /root/

COPY --from=builder /app/rating-scrapper .


RUN echo "0 * * * * /root/rating-scrapper >> /var/log/rating-scrapper.log 2>&1" > /etc/crontabs/root

RUN touch /var/log/rating-scrapper.log

RUN echo '#!/bin/sh' > /root/start.sh && \
    echo 'crond -f -d 8' >> /root/start.sh && \
    chmod +x /root/start.sh

CMD ["/root/start.sh"]