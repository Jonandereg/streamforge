
---

### `docs/adr/provider-finnhub.md`


# ADR Market Data Provider: Finnhub

**Status:** Accepted  
**Date:** 2025-08-27  

---

## Context
StreamForge requires a market data provider for fetching and ingesting financial ticks (equities/FX) in near-real time.  
The Project needed:
- Free or low-cost tier for development.  
- WebSocket or REST endpoints for live/delayed ticks.  
- Clear terms of service (ToS) and documented rate limits.  
- Sufficient metadata to normalize across providers.  

Providers considered:  
- IEX Cloud (free tier, equities only).  
- Alpha Vantage (equities/FX, slower API).  
- Twelve Data (trial tier, both equities/FX).  
- Finnhub (REST + WebSocket, generous free tier, simple schema).  

---

## Decision
I selected **Finnhub** as the initial provider.  

Reasons:  
- Free tier allows live/delayed equities and FX data.  
- WebSocket interface is simple and developer-friendly.  
- JSON payloads are minimal (`p,s,t,v`) and map cleanly to our normalized `Tick`.  
- Token-based auth via query parameter keeps setup trivial.  
- Community support and active documentation.

---

## Constraints
- **Rate limits (free tier):**
  - REST: 60 requests/minute.  
  - WebSocket: limited symbols and delayed trades without a paid plan.  
- **ToS:** Free tier data cannot be redistributed or used commercially. Intended for development and personal use.  
- **Data delay:** Real-time data requires a paid subscription.  

---

## Normalization Spec
All providers are mapped into a unified `Tick` struct:

| Field      | Type        | Finnhub source | Notes                             |
|------------|-------------|----------------|-----------------------------------|
| `symbol`   | string      | `s`            | e.g. `"AAPL"`, `"EURUSD"`.        |
| `ts`       | `time.Time` | `t` (epoch ms) | Converted to UTC.                 |
| `price`    | float64     | `p`            | Last trade/quote price.           |
| `size`     | float64     | `v`            | Trade size (0 if unknown).        |
| `exchange` | string      | `x` (optional) | Empty if not provided.            |
| `src_id`   | string      | constant       | `"finnhub"`.                      |

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
