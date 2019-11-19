DROP TABLE IF EXISTS post;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id int PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  email varchar(255) NOT NULL,
  imei varchar(255) NOT NULL,
  password varchar(255)
);

CREATE TABLE posts (
  id int PRIMARY KEY AUTO_INCREMENT,
  title varchar(255) NOT NULL,
  user_id int,
  path varchar(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
