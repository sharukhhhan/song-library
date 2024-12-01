package service

import (
	"context"
	"effective_mobile_tz/internal/entity"
	"effective_mobile_tz/internal/repository"
)

type Song interface {
	CreateSong(ctx context.Context, groupName, title string) (string, error)
	GetPaginatedLyrics(ctx context.Context, songID string, page, limit int) ([]entity.LyricsVerse, error)
	GetSongsByFilter(ctx context.Context, filter *entity.SongFilter) ([]entity.Song, error)
	GetSongByID(ctx context.Context, songID string) (*entity.Song, error)
	UpdateSong(ctx context.Context, update *entity.SongUpdate) error
	DeleteSong(ctx context.Context, songID string) error
}

type Service struct {
	Song
}

type Dependencies struct {
	Repository     *repository.Repository
	ExternalApiURL string
}

func NewService(dependencies Dependencies) *Service {
	return &Service{
		Song: NewSongService(
			dependencies.Repository.Song,
			dependencies.Repository.Group,
			dependencies.Repository.Lyrics,
			dependencies.Repository.DBTransaction,
			dependencies.ExternalApiURL),
	}
}
