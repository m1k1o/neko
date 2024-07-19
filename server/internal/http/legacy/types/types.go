package types

type Member struct {
	ID    string `json:"id"`
	Name  string `json:"displayname"`
	Admin bool   `json:"admin"`
	Muted bool   `json:"muted"`
}

type FileListItem struct {
	Filename string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
}

type ScreenConfiguration struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Rates  map[int]int16 `json:"rates"`
}
