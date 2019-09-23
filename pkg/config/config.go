package config

import (
    "github.com/BurntSushi/toml"
    //"github.com/davecgh/go-spew/spew"
    "github.com/fastly/fastly_lem/pkg/api"
    //"os"
)

type BigQueryConfig struct {
 Project    string `toml:"project"`
 Dataset    string `toml:"dataset"`
 Table      string `toml:"table"`
 Email      string `toml:"email"`
 PrivateKey string `toml:"private_key"`
}

type FastlyConfig struct {
    Snippets []SnippetConfig `toml:"snippet"`
    GlobalLoggingCondition string `toml:"global_logging_condition"`
    DictionaryName string `toml:"dictionary_name"`
    LoggingConfigName string `toml:"logging_config_name"`
}

type SnippetConfig struct {
    Name string `toml:"name"`
    Url string `toml:"url"`
    SnippetType string `toml:"type"`
    Priority int `toml:"priority"`
}

type LEMConfig struct {
   Fastly FastlyConfig `toml:"fastly"`
   BigQuery BigQueryConfig `toml:"bigquery"`
   Api *api.ApiClient `toml:"-"`
}

// New - Creates a new instance of the LEM Configuration
func New(configFile string, token string, serviceId string, version int) (LEMConfig, error) {
   var config LEMConfig

   if _, err := toml.DecodeFile(configFile,&config); err != nil {
    return LEMConfig{}, err
   }

   client, err := api.New(token,serviceId,version)
   if err != nil {
       return LEMConfig{}, err
   }

   config.Api = &client

   return config, nil
}

// SetupSnippets - Creates the necessary snippets in the config with the default priority
func (c *LEMConfig) SetupSnippets() error {
    //Download and create the snippets
    for _, s := range c.Fastly.Snippets {
        snippet, err := DownloadFile(s.Url)
        if err != nil {
            return err
        }

        if err = c.Api.CreateSnippet(s.Name,snippet,s.Priority,s.SnippetType); err != nil {
            return err
        }
    }

    return nil
}

// SetupCondition to disable global logging since logging will be done via
// the snippet itself
func (c *LEMConfig) SetupCondition() error {
   err := c.Api.CreateCondition(c.Fastly.GlobalLoggingCondition,"false",1,"response")
   return err
}

// SetupDictionary creates the new Edge Dictionary to conditionally control logging
func (c *LEMConfig) SetupDictionary() error {
    if err := c.Api.CreateDictionary(c.Fastly.DictionaryName); err != nil {
       return err
    }

    err := c.Api.CreateDictionaryItem(c.Fastly.DictionaryName,"enabled","0")
    return err
}

// SetupBigQuery - Creates the BigQuery configuration
func (c *LEMConfig) SetupBigQuery() error {
    err := c.Api.CreateBigQueryConfig(c.Fastly.LoggingConfigName,c.BigQuery.Project,c.BigQuery.Dataset, c.BigQuery.Table, c.BigQuery.Email, c.BigQuery.PrivateKey, c.Fastly.GlobalLoggingCondition)
    return err
}
