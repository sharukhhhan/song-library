package repository

import (
	"context"
	"effective_mobile_tz/internal/entity"
	"effective_mobile_tz/internal/repository/postgres"
	"github.com/jackc/pgx/v4"
)

type Song interface {
	CreateSong(ctx context.Context, song *entity.Song) (string, error)
	GetSongsByFilter(ctx context.Context, filter *entity.SongFilter) ([]entity.Song, error)
	DeleteSong(ctx context.Context, songID string) error
	GetSongByID(ctx context.Context, songID string) (*entity.Song, error)
	UpdateSong(ctx context.Context, update *entity.SongUpdate) error
}

type Group interface {
	GetGroupIDByName(ctx context.Context, name string) (string, error)
	CreateGroup(ctx context.Context, name string) (string, error)
}

type Lyrics interface {
	AddLyricsVerse(ctx context.Context, verse *entity.LyricsVerse) error
	GetAllLyrics(ctx context.Context, songID string) ([]entity.LyricsVerse, error)
	GetPaginatedLyrics(ctx context.Context, songID string, limit, offset int) ([]entity.LyricsVerse, error)
	DeleteLyrics(ctx context.Context, songID string) error
}

type DBTransaction interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	Song
	Group
	Lyrics
	DBTransaction
}

func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{
		Song:          postgres.NewSongPostgres(conn),
		Group:         postgres.NewGroupPostgres(conn),
		Lyrics:        postgres.NewLyricsPostgres(conn),
		DBTransaction: postgres.NewDBConn(conn),
	}
}
