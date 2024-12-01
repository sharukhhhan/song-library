package entity

import (
	"time"
)

type Song struct {
	ID          string    `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	GroupID     string    `db:"group_id" json:",omitempty"`
	GroupName   string    `json:"groupName"`
	ReleaseDate time.Time `db:"releaseDate" json:"releaseDate"`
	LyricsText  string    `json:"lyrics"`
	Link        string    `db:"link" json:"link"`
}

type SongUpdate struct {
	ID          string  `json:"id"`
	Title       *string `db:"title" json:"title"`
	ReleaseDate *string `db:"release_date" json:"releaseDate"`
	GroupName   *string `json:"groupName"`
	GroupID     string
	Link        *string `json:"link"`
	Lyrics      *string `json:"lyrics"`
}

type SongFilter struct {
	Title     string
	Link      string
	Group     string
	Text      string
	StartDate string
	EndDate   string
	Limit     int
	Offset    int
}
