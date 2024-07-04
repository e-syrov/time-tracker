CREATE TABLE users
(
    id              SERIAL PRIMARY KEY,
    surname         VARCHAR(100) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    patronymic      VARCHAR(100),
    passport_number VARCHAR(20)  NOT NULL UNIQUE,
    address         TEXT         NOT NULL
);

CREATE TABLE tasks
(
    task_id    SERIAL PRIMARY KEY,
    user_id    INT       NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time   TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id)
);
