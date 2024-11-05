package metadata

type EntryInfo struct {
	Name   string `json:"name"`
	Series string `json:"series"`

	IsManga        bool `json:"isManga"`
	PreferVertical bool `json:"preferVertical"`

	Cover string   `json:"cover"`
	Pages []string `json:"pages"`
}

type SeriesType string

const (
	SeriesTypeManga       SeriesType = "manga"
	SeriesTypeComic       SeriesType = "comic"
	SeriesTypeVisualNovel SeriesType = "visual_novel"
)

type SeriesInfoCover struct {
	Original string `json:"original"`
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
}

type SeriesInfo struct {
	Name string     `json:"name"`
	Type SeriesType `json:"type"`

	MalId     string `json:"malId"`
	AnilistId string `json:"anilistId"`

	Cover SeriesInfoCover `json:"cover"`
}
