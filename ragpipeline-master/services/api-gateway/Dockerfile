# API Gateway — multi-stage build
# Build context must be the repository root (docker-compose already does this):
#   docker build -f services/api-gateway/Dockerfile .
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Shared local modules referenced via replace directives
COPY pkg/ pkg/
COPY shared/ shared/
COPY services/api-gateway/ services/api-gateway/

WORKDIR /src/services/api-gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-gateway .

FROM alpine:3.20
RUN apk add --no-cache curl ca-certificates && adduser -D -u 10001 app
USER app
COPY --from=builder /out/api-gateway /usr/local/bin/api-gateway
EXPOSE 8080
ENTRYPOINT ["api-gateway"]
