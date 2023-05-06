CREATE TABLE users(
    id  SERIAL  PRIMARY KEY,
    username  TEXT NOT NULL,
    userslug  TEXT NOT NULL UNIQUE,
    email  TEXT NOT NULL UNIQUE,
    password  TEXT NOT NULL,
    joined TIMESTAMP NOT NULL
    --UNIQUE(userslug, email)
);