package config

import (
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fastly/fastly-lem/pkg/api"
)

// BigQueryConfig holds the configuration to the BigQuery logging object in the service config
type BigQueryConfig struct {
	Project    string `toml:"project"`
	Dataset    string `toml:"dataset"`
	Table      string `toml:"table"`
	Email      string `toml:"email"`
	PrivateKey string `toml:"private_key"`
}

// FastlyConfig holds the configuration to standard fastly configuration parameters, these should not be
// changed by the customer
type FastlyConfig struct {
	Snippets          []SnippetConfig `toml:"snippet"`
	DictionaryName    string          `toml:"dictionary_name"`
	LoggingConfigName string          `toml:"logging_config_name"`
	SampleRate        string          `toml:"sample_rate"`
	LogAllErrors      string          `toml:"log_all_errors"`
}

// SnippetConfig holds the list of snippets we will deploy to the service
type SnippetConfig struct {
	Name        string `toml:"name"`
	URL         string `toml:"url"`
	SnippetType string `toml:"type"`
	Priority    int    `toml:"priority"`
}

// Config is the parent configuration struct of the TOML config file
type Config struct {
	Fastly   FastlyConfig   `toml:"fastly"`
	BigQuery BigQueryConfig `toml:"bigquery"`
	API      *api.Client    `toml:"-"`
}

// New creates a new instance of the LEM Configuration
func New(configFile string, token string, serviceID string, version int) (Config, error) {
	var config Config

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return Config{}, err
	}

	client, err := api.NewClient(token, serviceID, version)
	if err != nil {
		return Config{}, err
	}

	config.API = &client

	return config, nil
}

// SetupSnippets creates the necessary snippets in the config with the default priority
func (c *Config) SetupSnippets() error {
	//Download and create the snippets
	for _, s := range c.Fastly.Snippets {
		snippet, err := DownloadFile(s.URL)
		if err != nil {
			return err
		}

		if err = c.API.CreateSnippet(s.Name, snippet, s.Priority, s.SnippetType); err != nil {
			return err
		}
	}

	return nil
}

// SetupDictionary creates the new Edge Dictionary to conditionally control logging
func (c *Config) SetupDictionary() error {
	if err := c.API.CreateDictionary(c.Fastly.DictionaryName); err != nil {
		return err
	}

	if err := c.API.CreateDictionaryItem(c.Fastly.DictionaryName, "enabled", "0"); err != nil {
		return err
	}

	if len(strings.TrimSpace(c.Fastly.LogAllErrors)) == 0 {
		c.Fastly.LogAllErrors = "0"
	}
	if err := c.API.CreateDictionaryItem(c.Fastly.DictionaryName, "log_all_errors", c.Fastly.LogAllErrors); err != nil {
		return err
	}

	if err := c.API.CreateDictionaryItem(c.Fastly.DictionaryName, "sample_rate", c.Fastly.SampleRate); err != nil {
		return err
	}

	return nil
}

// SetupBigQuery creates the BigQuery configuration
func (c *Config) SetupBigQuery() error {
	return c.API.CreateBigQueryConfig(c.Fastly.LoggingConfigName, c.BigQuery.Project, c.BigQuery.Dataset, c.BigQuery.Table, c.BigQuery.Email, c.BigQuery.PrivateKey)
}
