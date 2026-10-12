CREATE TABLE messages (
    id BIGINT NOT NULL AUTO_INCREMENT,

    type VARCHAR(16) NOT NULL,

    from_id BIGINT NOT NULL,
    from_nickname VARCHAR(64) NOT NULL,

    to_id BIGINT NULL,

    room_id BIGINT NULL,
    room_name VARCHAR(64) NULL,

    content TEXT NOT NULL,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    INDEX idx_room_time (room_id, created_at),
    INDEX idx_private_time (from_id, to_id, created_at)
);