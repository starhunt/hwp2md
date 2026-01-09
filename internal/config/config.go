// Package config manages application configuration.
package config

// Config represents the application configuration.
type Config struct {
	DefaultProvider string              `yaml:"default_provider"`
	Providers       map[string]Provider `yaml:"providers"`
	Format          FormatConfig        `yaml:"format"`
}

// Provider represents an LLM provider configuration.
type Provider struct {
	APIKey    string `yaml:"api_key"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
	Endpoint  string `yaml:"endpoint,omitempty"` // for Ollama or custom endpoints
}

// FormatConfig contains formatting options.
type FormatConfig struct {
	Temperature float64 `yaml:"temperature"`
	Language    string  `yaml:"language"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: "anthropic",
		Providers: map[string]Provider{
			"openai": {
				APIKey:    "${OPENAI_API_KEY}",
				Model:     "gpt-4o-mini",
				MaxTokens: 4096,
			},
			"anthropic": {
				APIKey:    "${ANTHROPIC_API_KEY}",
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 4096,
			},
			"gemini": {
				APIKey:    "${GOOGLE_API_KEY}",
				Model:     "gemini-1.5-flash",
				MaxTokens: 4096,
			},
			"ollama": {
				Endpoint:  "http://localhost:11434",
				Model:     "llama3.2",
				MaxTokens: 4096,
			},
		},
		Format: FormatConfig{
			Temperature: 0.3,
			Language:    "ko",
		},
	}
}

// GetProvider returns the provider configuration by name.
func (c *Config) GetProvider(name string) (*Provider, bool) {
	p, ok := c.Providers[name]
	if !ok {
		return nil, false
	}
	return &p, true
}

// GetDefaultProvider returns the default provider configuration.
func (c *Config) GetDefaultProvider() (*Provider, bool) {
	return c.GetProvider(c.DefaultProvider)
}
