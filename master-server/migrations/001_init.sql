CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(32) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS player_stats (
    user_id           INT PRIMARY KEY REFERENCES users(id),
    games_played      INT DEFAULT 0,
    games_won         INT DEFAULT 0,
    kills             INT DEFAULT 0,
    escapes           INT DEFAULT 0,
    generators_done   INT DEFAULT 0,
    survivors_hooked  INT DEFAULT 0,
    games_as_killer   INT DEFAULT 0,
    games_as_survivor INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS games (
    id           SERIAL PRIMARY KEY,
    started_at   TIMESTAMP DEFAULT NOW(),
    ended_at     TIMESTAMP,
    status       VARCHAR(20) DEFAULT 'in_progress',
    killer_id    INT REFERENCES users(id),
    container_id VARCHAR(64),
    port         INT,
    result       VARCHAR(32)
);

CREATE TABLE IF NOT EXISTS game_players (
    game_id    INT REFERENCES games(id),
    user_id    INT REFERENCES users(id),
    role       VARCHAR(10),
    survived   BOOLEAN DEFAULT FALSE,
    kills      INT DEFAULT 0,
    gens_done  INT DEFAULT 0,
    PRIMARY KEY (game_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_player_stats_wins ON player_stats(games_won DESC);
