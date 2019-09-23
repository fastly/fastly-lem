package config

import (
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
	Snippets               []SnippetConfig `toml:"snippet"`
	GlobalLoggingCondition string          `toml:"global_logging_condition"`
	DictionaryName         string          `toml:"dictionary_name"`
	LoggingConfigName      string          `toml:"logging_config_name"`
}

// SnippetConfig holds the list of snippets we will deploy to the service
type SnippetConfig struct {
	Name        string `toml:"name"`
	Url         string `toml:"url"`
	SnippetType string `toml:"type"`
	Priority    int    `toml:"priority"`
}

// Config is the parent configuration struct of the TOML config file
type Config struct {
	Fastly   FastlyConfig   `toml:"fastly"`
	BigQuery BigQueryConfig `toml:"bigquery"`
	Api      *api.Client    `toml:"-"`
}

// NewClient creates a new instance of the LEM Configuration
func New(configFile string, token string, serviceId string, version int) (Config, error) {
	var config Config

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return Config{}, err
	}

	client, err := api.NewClient(token, serviceId, version)
	if err != nil {
		return Config{}, err
	}

	config.Api = &client

	return config, nil
}

// SetupSnippets creates the necessary snippets in the config with the default priority
func (c *Config) SetupSnippets() error {
	//Download and create the snippets
	for _, s := range c.Fastly.Snippets {
		snippet, err := DownloadFile(s.Url)
		if err != nil {
			return err
		}

		if err = c.Api.CreateSnippet(s.Name, snippet, s.Priority, s.SnippetType); err != nil {
			return err
		}
	}

	return nil
}

// SetupCondition is a method to disable global logging since logging will be done via
// the snippet itself
func (c *Config) SetupCondition() error {
	return c.Api.CreateCondition(c.Fastly.GlobalLoggingCondition, "false", 1, "response")
}

// SetupDictionary creates the new Edge Dictionary to conditionally control logging
func (c *Config) SetupDictionary() error {
	if err := c.Api.CreateDictionary(c.Fastly.DictionaryName); err != nil {
		return err
	}

	err := c.Api.CreateDictionaryItem(c.Fastly.DictionaryName, "enabled", "0")
	return err
}

// SetupBigQuery creates the BigQuery configuration
func (c *Config) SetupBigQuery() error {
	return c.Api.CreateBigQueryConfig(c.Fastly.LoggingConfigName, c.BigQuery.Project, c.BigQuery.Dataset, c.BigQuery.Table, c.BigQuery.Email, c.BigQuery.PrivateKey, c.Fastly.GlobalLoggingCondition)
}
