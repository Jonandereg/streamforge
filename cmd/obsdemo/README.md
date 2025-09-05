# obsdemo

A small demo service that exercises the `internal/obs` observability package.

## Features
- Exposes `/metrics`, `/healthz`, `/readyz`, and `/debug/pprof/*`.
- Root handler (`/`) generates logs and spans with trace correlation.
- Sends traces to the local OpenTelemetry Collector → Jaeger.
- Metrics are scraped by Prometheus → Grafana.

## Usage

Start the infra stack (collector, Jaeger, Prometheus, Grafana):

```bash
make up
```
## Run the demo

```bash
go run ./cmd/obsdemo
```

## Test endpoints:

```bash
curl http://localhost:8080/         # logs + trace in Jaeger
curl http://localhost:8080/metrics  # Prometheus metrics
curl http://localhost:8080/healthz  # liveness
curl http://localhost:8080/readyz   # readiness
```

## Open UIs

- Jaeger → http://localhost:16686
- Prometheus → http://localhost:9090
- Grafana → http://localhost:3000

Notes:
Not shipped to production — for local testing & validation only.
Shows how observability works end-to-end before integrating into the real app.