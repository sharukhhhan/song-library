package postgres

import (
	"context"
	"effective_mobile_tz/internal/entity"
	"effective_mobile_tz/internal/repository/repoerrors"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"strings"
)

type SongPostgres struct {
	*pgx.Conn
}

func NewSongPostgres(conn *pgx.Conn) *SongPostgres {
	return &SongPostgres{Conn: conn}
}

func (s *SongPostgres) CreateSong(ctx context.Context, song *entity.Song) (string, error) {
	query := `
		INSERT INTO songs (title, group_id, release_date, link)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var songID string
	err := s.QueryRow(ctx, query, song.Title, song.GroupID, song.ReleaseDate, song.Link).Scan(&songID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return "", repoerrors.ErrAlreadyExists
			}
		}
		return "", err
	}

	return songID, nil
}

func (s *SongPostgres) GetSongsByFilter(ctx context.Context, filter *entity.SongFilter) ([]entity.Song, error) {
	baseQuery := `SELECT DISTINCT ON (s.id) s.id, s.release_date, g.name, s.title, s.link FROM songs s JOIN groups g ON s.group_id = g.id JOIN lyrics_verses l ON s.id = l.song_id`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.StartDate != "" && filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("release_date BETWEEN $%d AND $%d", argIndex, argIndex+1))
		args = append(args, filter.StartDate, filter.EndDate)
		argIndex += 2
	} else if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("release_date >= $%d", argIndex))
		args = append(args, filter.StartDate)
		argIndex++
	} else if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("release_date <= $%d", argIndex))
		args = append(args, filter.EndDate)
		argIndex++
	}

	if filter.Title != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Title+"%")
		argIndex++
	}

	if filter.Link != "" {
		conditions = append(conditions, fmt.Sprintf("link = $%d", argIndex))
		args = append(args, filter.Link)
		argIndex++
	}

	if filter.Group != "" {
		conditions = append(conditions, fmt.Sprintf("g.name ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Group+"%")
		argIndex++
	}

	if filter.Text != "" {
		conditions = append(conditions, fmt.Sprintf("l.verse ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Text+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY s.id, release_date DESC"
	if filter.Limit != 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset != 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := s.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query song: %w", err)
	}
	defer rows.Close()

	var songs []entity.Song
	for rows.Next() {
		var song entity.Song
		if err := rows.Scan(&song.ID, &song.ReleaseDate, &song.GroupName, &song.Title, &song.Link); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		songs = append(songs, song)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return songs, nil
}

func (s *SongPostgres) UpdateSong(ctx context.Context, update *entity.SongUpdate) error {
	baseQuery := `UPDATE songs SET `
	var updates []string
	var args []interface{}
	argIndex := 1

	if update.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, update.Title)
		argIndex++
	}
	if update.ReleaseDate != nil {
		updates = append(updates, fmt.Sprintf("release_date = $%d", argIndex))
		args = append(args, update.ReleaseDate)
		argIndex++
	}
	if update.GroupName != nil {
		updates = append(updates, fmt.Sprintf("group_id = $%d", argIndex))
		args = append(args, update.GroupID)
		argIndex++
	}
	if update.Link != nil {
		updates = append(updates, fmt.Sprintf("link = $%d", argIndex))
		args = append(args, update.Link)
		argIndex++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := baseQuery + strings.Join(updates, ", ") + fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, update.ID)

	result, err := s.Exec(ctx, query, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return repoerrors.ErrAlreadyExists
			}
		}
		return fmt.Errorf("failed to update song with ID %s: %w", update.ID, err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected < 1 {
		return repoerrors.ErrNotFound
	}

	return nil

}

func (s *SongPostgres) DeleteSong(ctx context.Context, songID string) error {
	query := `DELETE FROM songs WHERE id = $1`

	result, err := s.Exec(ctx, query, songID)
	if err != nil {
		return fmt.Errorf("failed to delete song with ID %s: %w", songID, err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected < 1 {
		return repoerrors.ErrNotFound
	}

	return nil
}

func (s *SongPostgres) GetSongByID(ctx context.Context, songID string) (*entity.Song, error) {
	query := `
		SELECT
			s.id,
			s.title,
			s.release_date,
			g.name AS group_name,
			s.link
		FROM songs s
		JOIN groups g ON s.group_id = g.id
		WHERE s.id = $1
	`

	var song entity.Song
	err := s.QueryRow(ctx, query, songID).Scan(
		&song.ID,
		&song.Title,
		&song.ReleaseDate,
		&song.GroupName,
		&song.Link,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch the song: %w", err)
	}

	return &song, nil
}
