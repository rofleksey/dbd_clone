package db

import (
	"database/sql"
	"dbd-master/internal/models"
)

// User queries

func CreateUser(username, passwordHash string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at",
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Create initial stats
	_, err = DB.Exec("INSERT INTO player_stats (user_id) VALUES ($1)", user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Stats queries

func GetPlayerStats(userID int) (*models.PlayerStats, error) {
	stats := &models.PlayerStats{}
	err := DB.QueryRow(
		`SELECT ps.user_id, u.username, ps.games_played, ps.games_won, ps.kills, ps.escapes,
		 ps.generators_done, ps.survivors_hooked, ps.games_as_killer, ps.games_as_survivor
		 FROM player_stats ps JOIN users u ON ps.user_id = u.id WHERE ps.user_id = $1`,
		userID,
	).Scan(&stats.UserID, &stats.Username, &stats.GamesPlayed, &stats.GamesWon,
		&stats.Kills, &stats.Escapes, &stats.GeneratorsDone, &stats.SurvivorsHooked,
		&stats.GamesAsKiller, &stats.GamesAsSurvivor)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func GetStatsByUsername(username string) (*models.PlayerStats, error) {
	stats := &models.PlayerStats{}
	err := DB.QueryRow(
		`SELECT ps.user_id, u.username, ps.games_played, ps.games_won, ps.kills, ps.escapes,
		 ps.generators_done, ps.survivors_hooked, ps.games_as_killer, ps.games_as_survivor
		 FROM player_stats ps JOIN users u ON ps.user_id = u.id WHERE u.username = $1`,
		username,
	).Scan(&stats.UserID, &stats.Username, &stats.GamesPlayed, &stats.GamesWon,
		&stats.Kills, &stats.Escapes, &stats.GeneratorsDone, &stats.SurvivorsHooked,
		&stats.GamesAsKiller, &stats.GamesAsSurvivor)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func GetLeaderboard(limit int) ([]models.PlayerStats, error) {
	rows, err := DB.Query(
		`SELECT ps.user_id, u.username, ps.games_played, ps.games_won, ps.kills, ps.escapes,
		 ps.generators_done, ps.survivors_hooked, ps.games_as_killer, ps.games_as_survivor
		 FROM player_stats ps JOIN users u ON ps.user_id = u.id
		 ORDER BY ps.games_won DESC, ps.games_played DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PlayerStats
	for rows.Next() {
		s := models.PlayerStats{}
		err := rows.Scan(&s.UserID, &s.Username, &s.GamesPlayed, &s.GamesWon,
			&s.Kills, &s.Escapes, &s.GeneratorsDone, &s.SurvivorsHooked,
			&s.GamesAsKiller, &s.GamesAsSurvivor)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// Game queries

func CreateGame(killerID int, containerID string, port int) (*models.Game, error) {
	game := &models.Game{}
	err := DB.QueryRow(
		`INSERT INTO games (killer_id, container_id, port, status)
		 VALUES ($1, $2, $3, 'in_progress') RETURNING id, started_at, status, killer_id, port`,
		killerID, containerID, port,
	).Scan(&game.ID, &game.StartedAt, &game.Status, &game.KillerID, &game.Port)
	if err != nil {
		return nil, err
	}
	return game, nil
}

func UpdateGamePortAndContainer(gameID int, port int, containerID string) error {
	_, err := DB.Exec(
		"UPDATE games SET port = $1, container_id = $2 WHERE id = $3",
		port, containerID, gameID,
	)
	return err
}

func AddGamePlayer(gameID, userID int, role string) error {
	_, err := DB.Exec(
		"INSERT INTO game_players (game_id, user_id, role) VALUES ($1, $2, $3)",
		gameID, userID, role,
	)
	return err
}

func EndGame(gameID int, result string) error {
	_, err := DB.Exec(
		"UPDATE games SET status = 'completed', ended_at = NOW(), result = $1 WHERE id = $2",
		result, gameID,
	)
	return err
}

func UpdateGamePlayer(gameID, userID int, survived bool, kills, gensDone int) error {
	_, err := DB.Exec(
		`UPDATE game_players SET survived = $1, kills = $2, gens_done = $3
		 WHERE game_id = $4 AND user_id = $5`,
		survived, kills, gensDone, gameID, userID,
	)
	return err
}

func UpdatePlayerStats(report models.PlayerReport, won bool) error {
	wonInt := 0
	if won {
		wonInt = 1
	}

	escapedInt := 0
	if report.Survived {
		escapedInt = 1
	}

	killerInt := 0
	survivorInt := 0
	if report.Role == "killer" {
		killerInt = 1
	} else {
		survivorInt = 1
	}

	_, err := DB.Exec(
		`UPDATE player_stats SET
		 games_played = games_played + 1,
		 games_won = games_won + $1,
		 kills = kills + $2,
		 escapes = escapes + $3,
		 generators_done = generators_done + $4,
		 survivors_hooked = survivors_hooked + $5,
		 games_as_killer = games_as_killer + $6,
		 games_as_survivor = games_as_survivor + $7
		 WHERE user_id = $8`,
		wonInt, report.Kills, escapedInt, report.GensDone, report.Kills,
		killerInt, survivorInt, report.UserID,
	)
	return err
}

func GetActiveGames() ([]models.Game, error) {
	rows, err := DB.Query(
		`SELECT g.id, g.started_at, g.status, g.killer_id, u.username, g.port
		 FROM games g JOIN users u ON g.killer_id = u.id
		 WHERE g.status = 'in_progress' ORDER BY g.started_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		g := models.Game{}
		err := rows.Scan(&g.ID, &g.StartedAt, &g.Status, &g.KillerID, &g.KillerName, &g.Port)
		if err != nil {
			return nil, err
		}

		// Get players for this game
		playerRows, err := DB.Query(
			`SELECT gp.user_id, u.username, gp.role FROM game_players gp
			 JOIN users u ON gp.user_id = u.id WHERE gp.game_id = $1`,
			g.ID,
		)
		if err == nil {
			for playerRows.Next() {
				p := models.GamePlayer{}
				playerRows.Scan(&p.UserID, &p.Username, &p.Role)
				g.Players = append(g.Players, p)
			}
			playerRows.Close()
		}

		games = append(games, g)
	}
	return games, nil
}

func CancelGame(gameID int) error {
	_, err := DB.Exec(
		"UPDATE games SET status = 'cancelled', ended_at = NOW(), result = 'disconnected' WHERE id = $1",
		gameID,
	)
	return err
}

func GetGameByID(gameID int) (*models.Game, error) {
	game := &models.Game{}
	err := DB.QueryRow(
		`SELECT g.id, g.started_at, g.status, g.killer_id, g.container_id, g.port, COALESCE(g.result, '')
		 FROM games g WHERE g.id = $1`,
		gameID,
	).Scan(&game.ID, &game.StartedAt, &game.Status, &game.KillerID, &game.ContainerID, &game.Port, &game.Result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return game, nil
}
