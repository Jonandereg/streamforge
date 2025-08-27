## Finnhub WS Spike (local only)

A tiny probe to verify Finnhub WebSocket access and our normalization.  
**Not part of production builds** .

### Run

```bash
export FINNHUB_TOKEN=your_real_token
export FINNHUB_BASE_URL=wss://ws.finnhub.io

# A couple of symbols
go run ./cmd/spike-ws -symbols AAPL,MSFT