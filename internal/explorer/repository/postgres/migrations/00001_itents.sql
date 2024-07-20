-- +goose Up
-- +goose StatementBegin
CREATE TABLE intents (
    id UUID PRIMARY KEY,
    repository TEXT NOT NULL,
    since TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE intents;
-- +goose StatementEnd
