package filetransfer

const PluginName = "filetransfer"

const (
	FILETRANSFER_UPDATE = "filetransfer/update"
)

type Message struct {
	Enabled bool   `json:"enabled"`
	RootDir string `json:"root_dir"`
	Files   []Item `json:"files"`
}

type ItemType string

const (
	ItemTypeFile ItemType = "file"
	ItemTypeDir  ItemType = "dir"
)

type Item struct {
	Name string   `json:"name"`
	Type ItemType `json:"type"`
	Size int64    `json:"size,omitempty"`
}
