DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id int PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  email varchar(255) NOT NULL,
  imei varchar(255) NOT NULL,
  password varchar(255)
);
