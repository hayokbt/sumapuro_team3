CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_username (username),
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS user_scores (
    user_id BIGINT UNSIGNED NOT NULL,
    wins INT UNSIGNED NOT NULL DEFAULT 0,
    losses INT UNSIGNED NOT NULL DEFAULT 0,
    draws INT UNSIGNED NOT NULL DEFAULT 0,
    rating_score DECIMAL(10, 2) NOT NULL DEFAULT 1000.00,
    PRIMARY KEY (user_id),
    CONSTRAINT fk_user_scores_user
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS pending_registrations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    verification_code_hash CHAR(64) NOT NULL,
    verification_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0,
    expires_at DATETIME NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_pending_registrations_username (username),
    UNIQUE KEY uq_pending_registrations_email (email),
    KEY idx_pending_registrations_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS account_recoveries (
    user_id BIGINT UNSIGNED NOT NULL,
    verification_code_hash CHAR(64) NOT NULL,
    verification_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0,
    expires_at DATETIME NOT NULL,
    reset_token_hash CHAR(64) NULL,
    reset_expires_at DATETIME NULL,
    PRIMARY KEY (user_id),
    CONSTRAINT fk_account_recoveries_user
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE,
    KEY idx_account_recoveries_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
