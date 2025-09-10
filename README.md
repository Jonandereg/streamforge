# StreamForge

StreamForge is a Go-based project that simulates a production-grade **real-time market data pipeline**.
It ingests tick-level events through **Kafka**, processes and stores them in **TimescaleDB** for time-series queries, and uses **Redis** for caching and fast lookups.

To make it closer to real-world systems, the stack includes a full observability layer with **Prometheus**, **Grafana**, and **OpenTelemetry**, plus tracing through **Jaeger**.

This project is designed as a **learning environment** to practice building scalable, event-driven backends with modern tools, while keeping the structure, tooling, and CI/CD setup aligned with professional engineering practices.

---

## Quick Demo: Ingestor

The ingestor service streams real-time ticks from [Finnhub](https://finnhub.io/) into Kafka.

### 1. Set environment
```bash
cp .env.example .env
# edit .env with your FINNHUB_TOKEN
```

### 2. Start infra (Kafka, Timescale, Redis, Prometheus, Grafana, Jaeger)
```bash
make up
```

### 3. Run the ingestor
```bash
go run ./cmd/ingestor
```

### 4. Verify ticks are flowing
Consume from Kafka:
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-console-consumer.sh   --bootstrap-server localhost:29092   --topic ticks   --from-beginning   --property print.headers=true
```

Check metrics:
```bash
curl -s localhost:2112/metrics | grep ingestor_
```

You should see `ingestor_fetch_total` and `ingestor_publish_total` increasing.

---

## Architecture

```text
[Finnhub WS] ---> [Ingestor (Go)] ---> [Kafka: ticks topic] ---> [Consumers / TimescaleDB]
                       |
                       +--> [/metrics] Prometheus --> Grafana
                       +--> Jaeger traces
```

---

## Features

- **Go backend** with clean project structure, CI/CD, and linting.
- **Real-time ingestion** via **Apache Kafka**, streaming normalized tick events.
- **Time-series storage** using **TimescaleDB** for efficient queries.
- **In-memory caching** layer with **Redis** for low-latency lookups.
- **Observability stack**: Prometheus, Grafana, OpenTelemetry, and Jaeger for metrics, dashboards, and tracing.
- **Local dev environment** via Docker Compose with reproducible setup.
- **Production-style practices**: Makefile tasks, pinned tool versions, GitHub Actions CI.

---

## Getting Started

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.24+](https://go.dev/dl/)

### Clone and setup
```bash
git clone https://github.com/jonandereg/streamforge.git
cd streamforge
make tools      # install linters
make deps       # tidy dependencies
```

### Environment
Copy defaults and adjust as needed:
```bash
cp .env.example .env
```

---

## Observability

StreamForge includes a production-grade observability baseline out of the box:

- **Structured logging** with [zap](https://github.com/uber-go/zap), JSON or console, sampling, caller info, and trace correlation (`trace_id` / `span_id`).
- **Metrics** via Prometheus client: Go runtime + process collectors, custom registry, and a `/metrics` endpoint (OpenMetrics enabled).
- **Tracing** with OpenTelemetry SDK → OTel Collector → Jaeger, including service metadata and configurable sampling.
- **Health endpoints** (`/healthz`, `/readyz`) and optional **pprof** (`/debug/pprof/*`).
- **Graceful shutdown** flushing logs and traces.

### Demo service

A lightweight demo app in [`cmd/obsdemo`](./cmd/obsdemo) exercises the observability stack end-to-end:

- Exposes `/`, `/metrics`, `/healthz`, `/readyz`, `/debug/pprof/*`.
- Sends traces through the OTel Collector into Jaeger.
- Prometheus scrapes metrics, visualized in Grafana.

See [`cmd/obsdemo/README.md`](./cmd/obsdemo/README.md) for instructions on running the demo locally with Docker Compose.

---

## Local Stack (Docker Compose, KRaft Kafka)

This repo ships a local infra stack for development:

- TimescaleDB (Postgres + time-series)
- Redis
- Kafka **in KRaft mode** (no ZooKeeper) via Bitnami image
- Prometheus (metrics), Grafana (dashboards)
- OpenTelemetry Collector (OTLP → Jaeger), Jaeger (traces)
- `DATABASE_URL` is used by the migration tool; it is generated from the Timescale env vars in local/dev.

### Database & Migrations

- **TimescaleDB** with `ticks` hypertable on `ts`.
- Dev retention: **30 days**; compression on chunks older than **7 days**.
- Managed with **golang-migrate** (via Docker).

Run migrations:
```bash
make migrate-up        # apply all up migrations
make migrate-version   # show current migration version
make migrate-down      # roll back one migration
```

---

## Normalized `Tick` Model

All providers are normalized into a common `Tick` struct before persistence, making the pipeline provider-agnostic.

| Field      | Type        | Source (Finnhub) | Notes                                |
|------------|-------------|------------------|--------------------------------------|
| `symbol`   | string      | `s`              | e.g. "AAPL", "EURUSD".               |
| `ts`       | `time.Time` | `t` (epoch ms)   | Converted to UTC.                    |
| `price`    | float64     | `p`              | Last trade/quote price.              |
| `size`     | float64     | `v`              | Trade size/volume (0 if unknown).    |
| `exchange` | string      | `x` (optional)   | Empty string if not provided.        |
| `src_id`   | string      | constant         | "finnhub" (provider ID).             |

Example normalized tick:
```json
{
  "symbol": "AAPL",
  "ts": "2025-08-27T20:15:31.134Z",
  "price": 231.42,
  "size": 100,
  "exchange": "",
  "src_id": "finnhub"
}
```

---

## Accessing Services

- TimescaleDB → `localhost:5432` (user: postgres, password: postgres, db: streamforge)
- Kafka → `localhost:29092` (external listener)
- Redis → `localhost:6379`
- Prometheus → http://localhost:9090
- Grafana → http://localhost:3000 (user: admin, password: admin)
- Jaeger UI → http://localhost:16686

---

## Kafka Quickstart

Create a test topic:
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-topics.sh   --bootstrap-server kafka:9092   --create --topic test-topic --partitions 3 --replication-factor 1
```

List topics:
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-topics.sh   --bootstrap-server kafka:9092 --list
```

Produce messages:
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-console-producer.sh   --bootstrap-server kafka:9092 --topic test-topic
```

Consume messages:
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-console-consumer.sh   --bootstrap-server kafka:9092 --topic test-topic   --group sf-demo-$(date +%s) --from-beginning --timeout-ms 10000
```

---

## Project Status

✅ Ingestor service runs, publishes ticks into Kafka, exposes Prometheus metrics.  
🚧 Next: consumers & TimescaleDB persistence.