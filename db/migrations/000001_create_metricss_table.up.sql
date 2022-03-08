BEGIN;

CREATE TYPE enum_types AS ENUM (
    'counter',
    'gauge'
);

CREATE TABLE IF NOT EXISTS metrics(
    metric_id VARCHAR (50) PRIMARY KEY,
    metric_type enum_types,
    metric_delta BIGINT,
    metric_value DOUBLE PRECISION
);

COMMIT;