CREATE TABLE IF NOT EXISTS readings (
    station SYMBOL,
    sensor SYMBOL,
    value DOUBLE,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp);
