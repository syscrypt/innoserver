DROP VIEW IF EXISTS detailed_posts;
DROP TABLE IF EXISTS options;
DROP TABLE IF EXISTS group_user;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id int PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  email varchar(255) NOT NULL UNIQUE,
  imei varchar(255) NOT NULL,
  password varchar(255)
);

CREATE TABLE groups (
  id int PRIMARY KEY AUTO_INCREMENT,
  title varchar(255) NOT NULL,
  admin_id int NOT NULL,
  unique_id varchar(255) NOT NULL UNIQUE,
  public tinyint(1) DEFAULT 0,
  FOREIGN KEY(admin_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE posts (
  id int PRIMARY KEY AUTO_INCREMENT,
  title varchar(255) NOT NULL,
  user_id int NOT NULL,
  path varchar(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  unique_id varchar(255) NOT NULL UNIQUE,
  parent_id int DEFAULT NULL,
  method int NOT NULL,
  type tinyint(1) NOT NULL,
  group_id int DEFAULT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY(parent_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE group_user (
  group_id int NOT NULL,
  user_id int NOT NULL,
  FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE options (
  post_uid varchar(255) NOT NULL,
  opt_key varchar(255) NOT NULL,
  opt_value varchar(255) NOT NULL,
  FOREIGN KEY(post_uid) REFERENCES posts(unique_id) ON DELETE CASCADE
);

CREATE VIEW detailed_posts AS
  SELECT a.*, COALESCE(b.unique_id, "") AS parent_uid,
         COALESCE(d.unique_id, "") AS group_uid,
    c.name FROM posts a LEFT JOIN posts b ON a.parent_id = b.id
                        LEFT JOIN users c on a.user_id = c.id
                        LEFT JOIN groups d on a.group_id = d.id;
