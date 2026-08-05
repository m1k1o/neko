package filetransfer

const PluginName = "filetransfer"

type Settings struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
}

const (
	FILETRANSFER_UPDATE = "filetransfer/update"
)

type Message struct {
	Enabled      bool   `json:"enabled"`
	RootDir      string `json:"root_dir"`
	UserDownload bool   `json:"user_download"`
	UserUpload   bool   `json:"user_upload"`
	Files        []Item `json:"files"`
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
