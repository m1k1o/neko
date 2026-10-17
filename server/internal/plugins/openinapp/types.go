package openinapp

const PluginName = "openinapp"

const (
	OPENINAPP_INIT    = "openinapp/init"
	OPENINAPP_OPENLINK = "openinapp/openlink"
)

type Init struct {
	Enabled bool `json:"enabled"`
}

type Url struct {
	Text string `json:"text"`
}
