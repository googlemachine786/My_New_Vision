# Query Understanding Service — multi-stage build (context = repo root)
FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY pkg/ pkg/
COPY shared/ shared/
COPY services/query-understanding-service/ services/query-understanding-service/

WORKDIR /src/services/query-understanding-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/query-understanding-service .

FROM alpine:3.20
RUN apk add --no-cache curl ca-certificates && adduser -D -u 10001 app
USER app
COPY --from=builder /out/query-understanding-service /usr/local/bin/query-understanding-service
EXPOSE 8083
ENTRYPOINT ["query-understanding-service"]
