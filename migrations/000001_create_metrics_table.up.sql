CREATE TABLE metrics
(
    id character varying(128) NOT NULL,
    mtype character varying(16) NOT NULL,
    value double precision,
    delta bigint,
    PRIMARY KEY (id)
);

CREATE INDEX idx_metrics_id_mtype ON metrics(id, mtype);