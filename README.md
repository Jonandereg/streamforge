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
git clone https://github.com/<your-username>/streamforge.git
cd streamforge
make tools      # install linters
make deps       # tidy dependencies
```

### Run the stack
```bash
make up
make ps
```

### Access services
- TimescaleDB → localhost:5432 (user: postgres, password: postgres, db: streamforge)
- Kafka → localhost:29092 (external listener)
- Redis → localhost:6379
- Prometheus → http://localhost:9090
- Grafana → http://localhost:3000 (user: admin, password: admin)
- Jaeger UI → http://localhost:16686

### Tear down
```bash
make down
```

### Status
Status

✅ Repository initialized, CI/CD pipeline green.

🚧 Currently implementing core ingestion pipeline.