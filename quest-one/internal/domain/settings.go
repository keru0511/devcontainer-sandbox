package domain

// Settings holds user-level application configuration.
type Settings struct {
	// ServerPort is the port the HTTP server listens on (default: 7890).
	ServerPort int

	// DataDir is the path to the directory storing the SQLite database.
	DataDir string

	// DatabaseFilename is the name of the SQLite database file.
	DatabaseFilename string

	// EncryptionEnabled indicates whether SQLCipher encryption is active.
	EncryptionEnabled bool

	// Language is the UI language code (e.g., "en", "ja").
	Language string

	// MaxCandidates is the number of tasks returned by the "candidates" command.
	MaxCandidates int

	// LLMProvider is the AI provider used for priority suggestions ("anthropic", "openai", "none").
	LLMProvider string

	// LLMModel is the specific model name for the LLM provider.
	LLMModel string

	// AutoSyncIntervalMinutes is how often background sync runs (0 = disabled).
	AutoSyncIntervalMinutes int

	// EnableMCP enables the MCP server mode.
	EnableMCP bool

	// MCPPort is the port for the HTTP-to-MCP proxy (default: 7891).
	MCPPort int
}

// DefaultSettings returns a Settings struct with sensible defaults.
func DefaultSettings() Settings {
	return Settings{
		ServerPort:              7890,
		DataDir:                 "~/.quest-one",
		DatabaseFilename:        "tasks.db",
		EncryptionEnabled:       false,
		Language:                "en",
		MaxCandidates:           5,
		LLMProvider:             "none",
		LLMModel:                "",
		AutoSyncIntervalMinutes: 60,
		EnableMCP:               false,
		MCPPort:                 7891,
	}
}

// Validate returns an error string if any setting is invalid.
func (s Settings) Validate() []string {
	var errs []string
	if s.ServerPort < 1 || s.ServerPort > 65535 {
		errs = append(errs, "server_port must be between 1 and 65535")
	}
	if s.DataDir == "" {
		errs = append(errs, "data_dir must not be empty")
	}
	if s.DatabaseFilename == "" {
		errs = append(errs, "database_filename must not be empty")
	}
	if s.MaxCandidates < 1 {
		errs = append(errs, "max_candidates must be at least 1")
	}
	return errs
}
