-- +goose Up
CREATE TABLE user_passkeys (
    credential_id bytea NOT NULL,
    user_id text NOT NULL,
    data jsonb NOT NULL,
    CONSTRAINT pk_user_passkeys PRIMARY KEY (credential_id),
    CONSTRAINT fk_user_passkeys_users_user_id
    FOREIGN KEY (user_id)
    REFERENCES users (id)
    ON DELETE CASCADE
);

CREATE INDEX idx_user_passkeys_user_id ON user_passkeys(user_id);

-- +goose Down
DROP TABLE user_passkeys;
