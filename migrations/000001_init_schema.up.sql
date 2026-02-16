CREATE TYPE role_type AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users (
	name TEXT NOT NULL UNIQUE PRIMARY KEY,
	hash_password TEXT NOT NULL,
	email TEXT NOT NULL,
	role role_type NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title TEXT NOT NULL,
	description TEXT,
	start_time TIMESTAMPTZ NOT NULL,
	end_time TIMESTAMPTZ NOT NULL,
	notify_before BIGINT,
	username TEXT references users(name) on delete cascade
);
