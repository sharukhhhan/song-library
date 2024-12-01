package entity

type LyricsVerse struct {
	ID          string `db:"id"`
	SongID      string `db:"song_id"`
	Verse       string `db:"verse"`
	VerseNumber int    `db:"verse_number"`
}
