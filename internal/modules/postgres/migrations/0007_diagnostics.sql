-- Migration 0007 — Diagnostics.
-- Tables: component health snapshots, support bundle records. Table only this
-- slice; the diagnostics surface and redacted support bundle land later
-- (Diagnostics and health).

CREATE TABLE IF NOT EXISTS component_health_snapshots (
    id         text        PRIMARY KEY,
    component  text        NOT NULL,
    state      text        NOT NULL,
    detail     text        NOT NULL DEFAULT '',
    checked_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS component_health_snapshots_component_idx
    ON component_health_snapshots (component, checked_at);

CREATE TABLE IF NOT EXISTS support_bundle_records (
    id         text        PRIMARY KEY,
    created_at timestamptz NOT NULL,
    location   text        NOT NULL DEFAULT '',
    redacted   boolean     NOT NULL DEFAULT true,
    size_bytes bigint      NOT NULL DEFAULT 0
);
