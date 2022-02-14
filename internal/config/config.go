package config

// Config struct
type Config struct {
	Port                    string `yaml:"port"`
	RedisURL                string `yaml:"redis_url"`
	RepoURL                 string `yaml:"repo_url"`
	ClientID                string `yaml:"client_id"`
	ClientSecret            string `yaml:"client_secret"`
	TokenExpiresTime        int    `yaml:"token_expires_time"`
	StreamerDataExpiresTime int    `yaml:"streamer_data_expires_time"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		Port:                    "8000",
		RedisURL:                "",
		RepoURL:                 "",
		ClientID:                "",
		ClientSecret:            "",
		TokenExpiresTime:        4320000,
		StreamerDataExpiresTime: 600,
	}
}
