package config

import (
    "github.com/BurntSushi/toml"
    "github.com/fastly/fastly_lem/pkg/api"
)

type BigQueryConfig struct {
 Project    string `toml:"project"`
 Dataset    string `toml:"dataset"`
 Table      string `toml:"table"`
 Email      string `toml:"email"`
 PrivateKey string `toml:"private_key"`
}

type SnippetConfig struct {
    Name string `toml:"name"`
    Url string `toml:"url"`
    SnippetType string `toml:"type"`
    Priority int `toml:"priority"`
}

type LEMConfig struct {
   Snippets []SnippetConfig `toml:"snippet"`
   BigQuery BigQueryConfig `toml:"bigquery"`
   GlobalLoggingCondition string `toml:"global_logging_condition"`
   DictionaryName string `toml:"dictionary_name"`
   LoggingConfigName string `toml:"logging_config_name"`
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
    for _, s := range c.Snippets {
        snippet, err := DownloadFile(s.Url)
        if err != nil {
            return err
        }

        err = c.Api.CreateSnippet(s.Name,snippet,s.Priority,s.SnippetType)
        if err != nil {
            return err
        }
    }

    return nil
}

// SetupCondition to disable global logging since logging will be done via
// the snippet itself
func (c *LEMConfig) SetupCondition() error {
   err := c.Api.CreateCondition(c.GlobalLoggingCondition,"false",1,"response")
   return err
}

// SetupDictionary creates the new Edge Dictionary to conditionally control logging
func (c *LEMConfig) SetupDictionary() error {
    err := c.Api.CreateDictionary(c.DictionaryName)
    if err != nil {
       return err
    }

    err = c.Api.CreateDictionaryItem(c.DictionaryName,"enabled","0")
    return err
}

// SetupBigQuery - Creates the BigQuery configuration
func (c *LEMConfig) SetupBigQuery() error {
    err := c.Api.CreateBigQueryConfig(c.LoggingConfigName,c.BigQuery.Project,c.BigQuery.Dataset, c.BigQuery.Table, c.BigQuery.Email, c.BigQuery.PrivateKey, c.GlobalLoggingCondition)
    return err
}