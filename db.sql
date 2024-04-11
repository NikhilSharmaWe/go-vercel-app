CREATE TABLE users(
	username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL PRIMARY KEY,
	github_access BOOLEAN DEFAULT FALSE
);

-- CREATE TABLE github_token (
--     username TEXT NOT NULL PRIMARY KEY REFERENCES users(username) ON DELETE CASCADE,
-- 	token TEXT NOT NULL
-- );