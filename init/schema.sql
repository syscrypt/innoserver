DROP TABLE IF EXISTS posts;
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
  unique_id varchar(255) NOT NULL UNIQUE,
  parent_uid varchar(255) NOT NULL,
  method int NOT NULL,
  type tinyint(1) NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
