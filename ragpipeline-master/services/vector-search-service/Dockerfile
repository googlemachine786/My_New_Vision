# Vector Search Service — multi-stage build (context = repo root)
FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY pkg/ pkg/
COPY shared/ shared/
COPY services/vector-search-service/ services/vector-search-service/

WORKDIR /src/services/vector-search-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/vector-search-service .

FROM alpine:3.20
RUN apk add --no-cache curl ca-certificates && adduser -D -u 10001 app
USER app
COPY --from=builder /out/vector-search-service /usr/local/bin/vector-search-service
EXPOSE 8082
ENTRYPOINT ["vector-search-service"]
