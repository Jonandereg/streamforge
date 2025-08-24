SELECT remove_compression_policy('ticks', if_exists => TRUE);
SELECT remove_retention_policy('ticks', if_exists => TRUE);


DROP INDEX IF EXISTS ticks_symbol_ts_desc_idx;
DROP INDEX IF EXISTS ticks_ts_idx;

-- Note: This call removes the hypertable metadata and underlying chunk tables.
SELECT drop_chunks('ticks', older_than => now());  
DROP TABLE IF EXISTS ticks CASCADE;
