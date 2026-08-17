CREATE TABLE users (
    id BIGINT NOT NULL AUTO_INCREMENT,

    username VARCHAR(64) COLLATE utf8mb4_unicode_ci NOT NULL,

    password_hash VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,

    nickname VARCHAR(64) COLLATE utf8mb4_unicode_ci NOT NULL,

    PRIMARY KEY (id),

    UNIQUE KEY username (username)

) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci;