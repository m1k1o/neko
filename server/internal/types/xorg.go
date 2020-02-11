package types

type ScreenSize struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rate   int16 `json:"rate"`
}

type ScreenConfiguration struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Rates  map[int]int16 `json:"rates"`
}
