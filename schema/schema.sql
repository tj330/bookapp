CREATE TABLE IF NOT EXISTS books (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255),
    description TEXT,
    author VARCHAR(255),
    isbn VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS ratings (
    record_id VARCHAR(255),
    record_type VARCHAR(255),
    user_id VARCHAR(255),
    value INTEGER,
    PRIMARY KEY (record_id, record_type, user_id)
);
