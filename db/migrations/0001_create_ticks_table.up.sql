CREATE EXTENSION IF NOT EXISTS timescaledb;


CREATE TABLE IF NOT EXISTS ticks (
  ts          timestamptz      NOT NULL,
  symbol      text             NOT NULL,
  price       numeric(18,6)    NOT NULL,
  size        numeric(18,6)    DEFAULT 0 NOT NULL,
  exchange    text             DEFAULT '' NOT NULL,
  src_id      text             NOT NULL,
  ingested_at timestamptz      NOT NULL DEFAULT now(),
  CONSTRAINT ticks_pk PRIMARY KEY (symbol, ts, src_id)
);

SELECT create_hypertable('ticks','ts', if_not_exists => TRUE, chunk_time_interval => INTERVAL '1 day');

CREATE INDEX IF NOT EXISTS ticks_symbol_ts_desc_idx ON ticks (symbol, ts DESC);
CREATE INDEX IF NOT EXISTS ticks_ts_idx ON ticks (ts);

SELECT add_retention_policy('ticks', INTERVAL '90 days', if_not_exists => TRUE);


ALTER TABLE ticks
  SET (
    timescaledb.compress = TRUE,
    timescaledb.compress_orderby = 'ts DESC',
    timescaledb.compress_segmentby = 'symbol'
  );


SELECT add_compression_policy('ticks', INTERVAL '7 days', if_not_exists => TRUE);