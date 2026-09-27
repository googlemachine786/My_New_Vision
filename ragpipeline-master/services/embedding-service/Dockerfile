# Embedding Service — multi-stage build (context = repo root)
FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY pkg/ pkg/
COPY shared/ shared/
COPY services/embedding-service/ services/embedding-service/

WORKDIR /src/services/embedding-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/embedding-service .

FROM alpine:3.20
RUN apk add --no-cache curl ca-certificates && adduser -D -u 10001 app
USER app
COPY --from=builder /out/embedding-service /usr/local/bin/embedding-service
EXPOSE 8081
ENTRYPOINT ["embedding-service"]
