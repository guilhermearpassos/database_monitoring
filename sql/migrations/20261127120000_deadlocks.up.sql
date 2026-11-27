CREATE TABLE IF NOT EXISTS deadlocks (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    f_id           uuid NOT NULL UNIQUE,
    target_id      int NOT NULL REFERENCES target (id),
    occurred_at    timestamp NOT NULL,
    xml_hash       varchar(64) NOT NULL,
    fingerprint    varchar(128) NOT NULL DEFAULT '',
    pattern        varchar(64) NOT NULL DEFAULT '',
    databases      text[] NOT NULL DEFAULT '{}',
    victim_spid    varchar(32) NOT NULL DEFAULT '',
    max_wait_ms    bigint NOT NULL DEFAULT 0,
    process_count  int NOT NULL DEFAULT 0,
    resource_count int NOT NULL DEFAULT 0,
    xml_report     text NOT NULL,
    CONSTRAINT uq_deadlocks_target_time_hash UNIQUE (target_id, occurred_at, xml_hash)
);

CREATE INDEX IF NOT EXISTS idx_deadlocks_target_occurred ON deadlocks (target_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_deadlocks_fingerprint ON deadlocks (target_id, fingerprint);
