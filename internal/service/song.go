package service

import (
	"context"
	"effective_mobile_tz/internal/entity"
	"effective_mobile_tz/internal/repository"
	"effective_mobile_tz/internal/repository/repoerrors"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SongService struct {
	songRepo      repository.Song
	groupRepo     repository.Group
	lyricsRepo    repository.Lyrics
	dbTransaction repository.DBTransaction
	externalAPI   string
}

func NewSongService(songPostgres repository.Song, groupPostgres repository.Group, lyricsRepo repository.Lyrics, dbTransaction repository.DBTransaction, externalAPI string) *SongService {
	return &SongService{
		songRepo:      songPostgres,
		groupRepo:     groupPostgres,
		lyricsRepo:    lyricsRepo,
		dbTransaction: dbTransaction,
		externalAPI:   externalAPI}
}

func (s *SongService) CreateSong(ctx context.Context, groupName, title string) (string, error) {
	// starting transaction
	tx, err := s.dbTransaction.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var groupID string
	groupID, err = s.groupRepo.GetGroupIDByName(ctx, groupName)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			groupID, err = s.groupRepo.CreateGroup(ctx, groupName)
			if err != nil {
				return "", fmt.Errorf("failed to create group: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to get group: %w", err)
		}
	}

	songDetail, err := fetchSongDetail(s.externalAPI, groupName, title)
	if err != nil {
		return "", fmt.Errorf("failed to access to external api: %w", err)
	}

	releaseDate, err := time.Parse("02.01.2006", songDetail.ReleaseDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse release date: %w", err)
	}
	song := &entity.Song{
		Title:       title,
		GroupID:     groupID,
		ReleaseDate: releaseDate,
		Link:        songDetail.Link,
	}

	songID, err := s.songRepo.CreateSong(ctx, song)
	if err != nil {
		if errors.Is(err, repoerrors.ErrAlreadyExists) {
			return "", ErrSongAlreadyExists
		}

		return "", err
	}

	if strings.Trim(songDetail.Text, " ") != "" {
		lyricsVerses := strings.Split(songDetail.Text, "\n")

		for verseNumber, verse := range lyricsVerses {
			lyricsVerse := &entity.LyricsVerse{
				SongID:      songID,
				Verse:       verse,
				VerseNumber: verseNumber + 1,
			}

			err = s.lyricsRepo.AddLyricsVerse(ctx, lyricsVerse)
			if err != nil {
				return "", fmt.Errorf("failed to add lyrics for the song: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return songID, nil
}

func (s *SongService) GetSongsByFilter(ctx context.Context, filter *entity.SongFilter) ([]entity.Song, error) {
	songs, err := s.songRepo.GetSongsByFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve songs: %w", err)
	}

	for ind, song := range songs {
		lyrics, err := s.lyricsRepo.GetAllLyrics(ctx, song.ID)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving lyrics for song: %w", err)
		}

		var lyricsSliceOfStrings []string
		for _, verse := range lyrics {
			lyricsSliceOfStrings = append(lyricsSliceOfStrings, verse.Verse)
		}

		song.LyricsText = strings.Join(lyricsSliceOfStrings, "\n")
		songs[ind] = song
	}

	return songs, nil
}

func (s *SongService) GetSongByID(ctx context.Context, songID string) (*entity.Song, error) {
	song, err := s.songRepo.GetSongByID(ctx, songID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return nil, ErrSongNotFound
		}

		return nil, fmt.Errorf("failed to retrieve the song: %w", err)
	}

	lyrics, err := s.lyricsRepo.GetAllLyrics(ctx, song.ID)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving lyrics for song: %w", err)
	}

	var lyricsSliceOfStrings []string
	for _, verse := range lyrics {
		lyricsSliceOfStrings = append(lyricsSliceOfStrings, verse.Verse)
	}

	song.LyricsText = strings.Join(lyricsSliceOfStrings, "\n")

	return song, nil
}

func (s *SongService) UpdateSong(ctx context.Context, update *entity.SongUpdate) error {
	// starting transaction
	tx, err := s.dbTransaction.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if update.GroupName != nil {
		update.GroupID, err = s.groupRepo.GetGroupIDByName(ctx, *update.GroupName)
		if err != nil {
			if errors.Is(err, repoerrors.ErrNotFound) {
				update.GroupID, err = s.groupRepo.CreateGroup(ctx, *update.GroupName)
				if err != nil {
					return fmt.Errorf("failed to create group: %w", err)
				}
			} else {
				return fmt.Errorf("failed to get group: %w", err)
			}
		}
	}

	err = s.songRepo.UpdateSong(ctx, update)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return ErrSongNotFound
		} else if errors.Is(err, repoerrors.ErrAlreadyExists) {
			return ErrSongAlreadyExists
		}

		return fmt.Errorf("failed to update the song: %w", err)
	}

	if update.Lyrics != nil {
		err = s.lyricsRepo.DeleteLyrics(ctx, update.ID)
		if err != nil {
			return fmt.Errorf("failed to delete old lyrics: %w", err)
		}

		lyricsVerses := strings.Split(*update.Lyrics, "\n")

		for verseNumber, verse := range lyricsVerses {
			lyricsVerse := &entity.LyricsVerse{
				SongID:      update.ID,
				Verse:       verse,
				VerseNumber: verseNumber + 1,
			}

			err = s.lyricsRepo.AddLyricsVerse(ctx, lyricsVerse)
			if err != nil {
				return fmt.Errorf("failed to add new lyrics for the song: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SongService) DeleteSong(ctx context.Context, songID string) error {
	// starting transaction
	tx, err := s.dbTransaction.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = s.lyricsRepo.DeleteLyrics(ctx, songID)
	if err != nil {
		return fmt.Errorf("failed to delete lyrics: %w", err)
	}

	err = s.songRepo.DeleteSong(ctx, songID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return ErrSongNotFound
		}
		return fmt.Errorf("failed to delete the song: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SongService) GetPaginatedLyrics(ctx context.Context, songID string, page, limit int) ([]entity.LyricsVerse, error) {
	offset := (page - 1) * limit
	return s.lyricsRepo.GetPaginatedLyrics(ctx, songID, limit, offset)
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func fetchSongDetail(apiURL, groupName, songTitle string) (*SongDetail, error) {
	params := url.Values{}
	params.Add("group", groupName)
	params.Add("song", songTitle)

	fullURL := fmt.Sprintf("%s/info?%s", apiURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		fmt.Printf("Response: %s", string(body))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var songDetail SongDetail
	if err := json.Unmarshal(body, &songDetail); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &songDetail, nil
}
