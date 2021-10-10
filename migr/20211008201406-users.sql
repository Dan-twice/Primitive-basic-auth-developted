
-- +migrate Up
CREATE TABLE users(
	user_id serial PRIMARY KEY,
	username VARCHAR(50) UNIQUE NOT NULL,
	password VARCHAR(100) NOT NULL,
	birth_date DATE NOT NULL,
	full_name VARCHAR(200) NOT NULL
);

INSERT INTO users (username, password, birth_date, full_name)
VALUES ('Raju', '4698c66370e53300c540e29c281bd695ee8657365258b0e31c9bbf940343a7c1', '1996-12-02', 'Raju Cutropaly');

-- +migrate Down
DROP TABLE users;
