CREATE TABLE IF NOT EXISTS user 
(
    id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    name TEXT,
    email VARCHAR(255),
    password VARCHAR(255),

    PRIMARY KEY (id),
    UNIQUE (id, email(255))
);

CREATE TABLE IF NOT EXISTS event (
    id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    title TEXT,
    start TIMESTAMP,
    end TIMESTAMP,
    tz TEXT,
    repeated TEXT,

    PRIMARY KEY (id),
    UNIQUE (id),
    CONSTRAINT fk_event_user FOREIGN KEY (user_id) REFERENCES user(id)
);