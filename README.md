# Market Data

A Go-based market data service for ingesting and processing real-time financial market data.

The project starts with Binance and is designed around a provider-agnostic architecture so additional exchanges and market data sources can be added without changing the core processing pipeline.

## What This Project Demonstrates

This repository focuses on practical backend engineering problems that appear in real-time and high-throughput systems:

- REST and WebSocket ingestion
- Concurrent data processing with goroutines and channels
- Context-based cancellation and graceful shutdown
- Backpressure and bounded concurrency
- Reconnection, retry, and failure handling
- Normalized market data models
- Metrics and observability
- CPU, memory, goroutine, blocking, and mutex profiling
- Runtime tracing
- Race detection
- Benchmarks and allocation analysis
- Testing concurrent Go code

## Architecture

The system is being developed incrementally around a small set of explicit responsibilities:

```text
Market Data Providers
        |
        |  REST / WebSocket
        v
+---------------------+
| Provider Adapters   |
+---------------------+
        |
        v
+---------------------+
| Normalization       |
+---------------------+
        |
        v
+---------------------+
| Processing Pipeline |
+---------------------+
        |
        +----> Metrics / Observability
        |
        +----> Consumers / Storage
```

The initial provider is Binance. Provider-specific protocol details are kept at the edge of the system; internal processing works with normalized domain models.

## Design Principles

- Keep goroutine ownership and lifecycle explicit.
- Prefer bounded concurrency over uncontrolled fan-out.
- Propagate cancellation with `context.Context`.
- Treat backpressure as a first-class design concern.
- Measure before optimizing.
- Prefer standard-library solutions when they are sufficient.
- Keep exchange-specific code isolated behind provider boundaries.
- Make failure behavior observable and testable.

## Planned Milestones

### 1. REST Market Data

- HTTP client
- timeouts and cancellation
- response validation
- JSON decoding
- typed errors

### 2. WebSocket Streaming

- persistent connections
- read loop
- heartbeat handling
- reconnect logic
- graceful shutdown

### 3. Concurrent Processing

- worker pools
- fan-out / fan-in
- bounded queues
- backpressure
- synchronization primitives

### 4. Reliability

- retry policies
- exponential backoff
- connection recovery
- signal handling
- failure isolation

### 5. Observability

- structured logging
- Prometheus metrics
- health endpoints
- runtime metrics

### 6. Performance Engineering

- `go test -bench`
- allocation analysis
- `pprof`
- `go tool trace`
- race detector
- goroutine leak analysis
- mutex and blocking profiles

## Markets

The core architecture is intended to support multiple data providers and market types over time, including:

- Crypto
- DeFi
- Forex
- Equities
- Futures
- Indices

## Technology

- Go 1.27+
- `net/http`
- WebSocket client library
- `context`
- Go concurrency primitives
- `runtime/pprof`
- `net/http/pprof`
- `go tool trace`
- Prometheus-compatible metrics

Additional dependencies will be introduced only where they provide a clear engineering benefit.

## Status

Early development.

The repository is intentionally being built in small, measurable increments. Features listed above are roadmap items unless they are already present in the source tree.

## Running

```bash
go run .
```

## Verification

As the project grows, changes should pass the relevant checks:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

Performance-sensitive changes should be supported by benchmarks rather than assumptions.
