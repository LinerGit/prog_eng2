-- +goose Up
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    current_node_id TEXT NOT NULL DEFAULT 'start',
    solution TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_tickets_status_created_at ON tickets (status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tickets_current_node_id ON tickets (current_node_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tickets_deleted_at ON tickets (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS tickets;
