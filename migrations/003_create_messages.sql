CREATE TABLE messages (
    id BIGINT NOT NULL AUTO_INCREMENT,

    message_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,

    type VARCHAR(16) NOT NULL,

    from_id BIGINT NOT NULL,
    from_nickname VARCHAR(64) NOT NULL,

    to_id BIGINT NULL,

    room_id BIGINT NULL,
    room_name VARCHAR(64) NULL,

    content TEXT NOT NULL,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),

    UNIQUE KEY uk_messages_message_id (message_id),

    INDEX idx_room_time (room_id, created_at),
    INDEX idx_private_time (from_id, to_id, created_at)
);
