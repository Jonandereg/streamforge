# StreamForge  

StreamForge is a Go-based project that simulates a production-grade **real-time market data pipeline**. It ingests tick-level events through **Kafka**, processes and stores them in **TimescaleDB** for time-series queries, and uses **Redis** for caching and fast lookups.  

To make it closer to real-world systems, the stack includes a full observability layer with **Prometheus**, **Grafana**, and **OpenTelemetry**, plus tracing through **Jaeger**.  

This project is designed as a **learning environment** to practice building scalable, event-driven backends with modern tools, while keeping the structure, tooling, and CI/CD setup aligned with professional engineering practices.  

---

## Features  

- **Go backend** with clean project structure, CI/CD, and linting.  
- **Real-time ingestion** via **Apache Kafka**, simulating market tick streams.  
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

## Local Stack (Docker Compose, KRaft Kafka)

This repo ships a local infra stack for development:

- TimescaleDB (Postgres + time-series)
- Redis
- Kafka **in KRaft mode** (no ZooKeeper) via Bitnami image
- Prometheus (metrics), Grafana (dashboards)
- OpenTelemetry Collector (OTLP → Jaeger), Jaeger (traces)

### Environment
Copy defaults and adjust as needed:

```bash
cp .env.example .env
```

### Run the stack
```bash
make up      # start stack
make ps      # show container status
make logs    # tail all container logs
make down    # stop & remove containers/volumes
```

### Access services
- TimescaleDB → localhost:5432 (user: postgres, password: postgres, db: streamforge)
- Kafka → localhost:29092 (external listener)
- Redis → localhost:6379
- Prometheus → http://localhost:9090
- Grafana → http://localhost:3000 (user: admin, password: admin)
- Jaeger UI → http://localhost:16686

### Kafka Quickstart

You can try Kafka right away with a test topic:

**Create a topic**
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server kafka:9092 \
  --create --topic test-topic --partitions 3 --replication-factor 1
  ```

**List topics**

```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server kafka:9092 --list
  ```

**Produce messages**
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server kafka:9092 --topic test-topic
  ```
Type some messages and press Enter to send.

**Consume messages**
```bash
docker exec -it sf-kafka /opt/bitnami/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server kafka:9092 --topic test-topic \
  --group sf-demo-$(date +%s) --from-beginning --timeout-ms 10000
  ```

### Tear down
```bash
make down
```

### Status
Status

✅ Repository initialized, CI/CD pipeline green.

🚧 Currently implementing core ingestion pipeline.