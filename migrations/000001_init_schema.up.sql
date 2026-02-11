CREATE TABLE IF NOT EXISTS users (
	name TEXT NOT NULL UNIQUE,
	hash_password TEXT,
	email TEXT,
	role TEXT,
	PRIMARY KEY (name)
);
CREATE INDEX IF NOT EXISTS idx_name ON users(name);

CREATE TABLE IF NOT EXISTS events (
	id serial primary key,
	title text,
	description text,
	start_time timestamptz,
	stop_time timestamptz,
	user_name TEXT references users(name) on delete cascade
);
