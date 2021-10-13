package config

// Config struct
type Config struct {
	Port     string `yaml:"port"`
	RedisURL string `yaml:"redis_url"`
	RepoURL  string `yaml:"repo_url"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		Port:     "8000",
		RedisURL: "",
		RepoURL:  "",
	}
}
