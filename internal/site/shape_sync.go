package site

// ShapeSyncConfig is optional site shape sync settings from .monms/config.json.
type ShapeSyncConfig struct {
	Enabled     bool   `json:"enabled"`
	Ref         string `json:"ref"`
	Remote      string `json:"remote"`
	Force       bool   `json:"force"`
	FailOnError bool   `json:"failOnError"`
}

// SyncOptions configures git fetch + checkout for shape promotion.
type SyncOptions struct {
	Ref    string
	Remote string
	Force  bool
}
