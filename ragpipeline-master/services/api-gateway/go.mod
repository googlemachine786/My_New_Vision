module github.com/visionary/ragpipeline/services/api-gateway

go 1.25.0

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/prometheus/client_golang v1.19.0
	github.com/redis/go-redis/v9 v9.5.0
	github.com/rs/zerolog v1.31.0
	github.com/visionary/ragpipeline/pkg/circuitbreaker v0.0.0-00010101000000-000000000000
	github.com/visionary/ragpipeline/pkg/contextkeys v0.0.0-00010101000000-000000000000
	github.com/visionary/ragpipeline/pkg/types v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	gonum.org/v1/gonum v0.15.0
)

replace (
	github.com/visionary/ragpipeline/pkg/circuitbreaker => ../../pkg/circuitbreaker
	github.com/visionary/ragpipeline/pkg/contextkeys => ../../pkg/contextkeys
	github.com/visionary/ragpipeline/pkg/httperrors => ../../pkg/httperrors
	github.com/visionary/ragpipeline/pkg/types => ../../pkg/types
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
