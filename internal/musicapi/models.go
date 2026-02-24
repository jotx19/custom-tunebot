package musicapi

type SongLite struct {
	ID        string
	Title     string
	Artist    string
	Image     string
	Link      string
	StreamURL string // direct audio stream if your API provides it
}

type SongDetail = SongLite
