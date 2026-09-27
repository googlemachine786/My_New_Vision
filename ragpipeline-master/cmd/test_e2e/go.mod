module github.com/visionary/ragpipeline/cmd/test_e2e

go 1.25

require (
	github.com/tmc/langchaingo v0.1.14
	github.com/visionary/ragpipeline/ingestion-go v0.0.0
	github.com/visionary/ragpipeline/orchestrator v0.0.0
)

replace github.com/visionary/ragpipeline/ingestion-go => ../../ingestion-go
replace github.com/visionary/ragpipeline/orchestrator => ../../orchestrator
