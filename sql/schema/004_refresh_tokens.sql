-- +goose Up
CREATE TABLE refresh_tokens(
	token TEXT PRIMARY KEY,
	revoked_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	user_id UUID NOT NULL,
	foreign KEY(user_id) REFERENCES users(id)
ON DELETE cascade
);

-- +goose Down
DROP TABLE refresh_tokens;
