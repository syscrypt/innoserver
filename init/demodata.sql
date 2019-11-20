SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE posts;
TRUNCATE TABLE users;

INSERT INTO users (name, email, imei, password) VALUES
("bob", "bob@bob.de", "12345678910", ""),
("alice", "alice@alice.de", "12345678910", ""),
("mallory", "mallory@mallory.de", "12345678910", "");

SET FOREIGN_KEY_CHECKS = 1;
