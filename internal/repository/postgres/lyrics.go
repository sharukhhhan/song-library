package postgres

import (
	"context"
	"effective_mobile_tz/internal/entity"
	"fmt"
	"github.com/jackc/pgx/v4"
)

type LyricsPostgres struct {
	*pgx.Conn
}

func NewLyricsPostgres(conn *pgx.Conn) *LyricsPostgres {
	return &LyricsPostgres{Conn: conn}
}

func (l *LyricsPostgres) AddLyricsVerse(ctx context.Context, verse *entity.LyricsVerse) error {
	query := `INSERT INTO lyrics_verses (song_id, verse, verse_number) VALUES ($1, $2, $3)`

	_, err := l.Exec(ctx, query, verse.SongID, verse.Verse, verse.VerseNumber)
	if err != nil {
		return err
	}

	return nil
}

func (l *LyricsPostgres) GetAllLyrics(ctx context.Context, songID string) ([]entity.LyricsVerse, error) {
	query := `SELECT verse_number, verse FROM lyrics_verses WHERE song_id = $1 ORDER BY verse_number;`

	rows, err := l.Query(ctx, query, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer rows.Close()

	var verses []entity.LyricsVerse

	for rows.Next() {
		var verse entity.LyricsVerse
		if err := rows.Scan(&verse.VerseNumber, &verse.Verse); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		verses = append(verses, verse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return verses, nil
}

func (l *LyricsPostgres) GetPaginatedLyrics(ctx context.Context, songID string, limit, offset int) ([]entity.LyricsVerse, error) {
	query := `
	SELECT verse_number, verse
	FROM lyrics_verses
	WHERE song_id = $1
	ORDER BY verse_number
	LIMIT $2 OFFSET $3;
`

	rows, err := l.Query(ctx, query, songID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer rows.Close()

	var lyrics []entity.LyricsVerse
	for rows.Next() {
		var verse entity.LyricsVerse
		if err := rows.Scan(&verse.VerseNumber, &verse.Verse); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		lyrics = append(lyrics, verse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return lyrics, nil
}

func (l *LyricsPostgres) DeleteLyrics(ctx context.Context, songID string) error {
	query := `DELETE FROM lyrics_verses WHERE song_id = $1`

	_, err := l.Exec(ctx, query, songID)
	if err != nil {
		return fmt.Errorf("failed to delete lyrics for song ID %s: %w", songID, err)
	}

	return nil
}
