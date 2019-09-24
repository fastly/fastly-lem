package api

import (
	"errors"
	"fmt"
	"github.com/fastly/go-fastly/fastly"
	"time"
)

// Client holds the information to perform updates via the Fastly API
type Client struct {
	Client    *fastly.Client
	ServiceID string
	Version   int
}

// DictionarySleepSeconds is our sleep timer waiting for a dictionary to be available
const DictionarySleepSeconds = 2

// NewClient creates a new instance of Client.
// Automatically set the version to the latest version we can
// write to unless it was supplied
func NewClient(key, id string, version int) (Client, error) {
	client, err := fastly.NewClient(key)
	if err != nil {
		return Client{}, err
	}

	if version == 0 {
		latest, err := GetLatestVersion(key, id)
		if err != nil {
			return Client{}, err
		}
		fmt.Printf("Using latest writeable version %d, since no version was provided\n", latest)
		version = latest
	}

	c := Client{
		client,
		id,
		version,
	}

	return c, nil
}

// GetLatestVersion figures out the latest version of the configuration
func GetLatestVersion(key string, serviceID string) (int, error) {
	client, err := fastly.NewClient(key)
	if err != nil {
		return 0, err
	}

	latest, err := client.LatestVersion(&fastly.LatestVersionInput{
		Service: serviceID,
	})
	if err != nil {
		return 0, err
	}

	// API returns nil if there is no version of the config
	if latest == nil {
		return 0, errors.New("there is currently no version in the service, please create a version in your service first")
	}

	// Check if latest version is not active
	if latest.Active == true {
		// TODO: see if we want to automatically clone for user?
		//newVersion, err := client.CloneVersion(&fastly.CloneVersionInput{
		//	Service: serviceID,
		//	Version: latest.Number,
		//})
		// fmt.Printf("cloning active version %d to new version %d since there is no draft config to write to",latest.Number, newVersion)
		return 0, fmt.Errorf("there is no writeable config version, the latest version %d is active and should be cloned before continuing", latest.Number)
	}

	return latest.Number, nil
}

// CreateSnippet creates a snippet in the service configuration
func (c *Client) CreateSnippet(name, content string, priority int, snippetType string) error {
	input := &fastly.CreateSnippetInput{
		Service:  c.ServiceID,
		Version:  c.Version,
		Name:     name,
		Priority: priority,
		Type:     toSnippetType(snippetType),
		Content:  content,
	}

	if _, err := c.Client.CreateSnippet(input); err != nil {
		return err
	}

	fmt.Printf("Snippet %s created for method %s in version %d\n", name, snippetType, c.Version)
	return nil
}

// CreateCondition creates a new condition in the API that we can attach to other objects
func (c *Client) CreateCondition(name, statement string, priority int, condType string) error {
	input := &fastly.CreateConditionInput{
		Service:   c.ServiceID,
		Version:   c.Version,
		Name:      name,
		Statement: statement,
		Priority:  priority,
		Type:      condType,
	}

	if _, err := c.Client.CreateCondition(input); err != nil {
		return err
	}

	fmt.Printf("Condition %s created with type %s in version %d\n", name, condType, c.Version)
	return nil
}

// CreateDictionary creates a new edge dictionary
func (c *Client) CreateDictionary(name string) error {
	input := &fastly.CreateDictionaryInput{
		Service: c.ServiceID,
		Version: c.Version,
		Name:    name,
	}

	if _, err := c.Client.CreateDictionary(input); err != nil {
		return err
	}

	fmt.Printf("Dictionary %s successfully created\n", name)
	return nil
}

// CheckDictionaryExists is a method to ensure the dictionary is available before writing new entries.
func (c *Client) CheckDictionaryExists(dictionary string) (string, bool) {
	input := &fastly.GetDictionaryInput{
		Service: c.ServiceID,
		Version: c.Version,
		Name:    dictionary,
	}

	d, err := c.Client.GetDictionary(input)

	if err != nil {
		return "", false
	}

	return d.ID, true
}

// CreateDictionaryItem creates a new element in the dictionary
func (c *Client) CreateDictionaryItem(dictionary, key, value string) error {
	//First let's loop up until dictionary is available via config
	var (
		d      string
		exists bool
	)

	deadline := time.Now().Add(60 * time.Second)
	fmt.Printf("Waiting for dictionary %s to be available.", dictionary)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for dictionary to be created")
		}

		if d, exists = c.CheckDictionaryExists(dictionary); !exists {
			fmt.Printf(".")
			time.Sleep(DictionarySleepSeconds * time.Second)
			continue
		}

		break
	}

	input := &fastly.CreateDictionaryItemInput{
		Service:    c.ServiceID,
		ItemKey:    key,
		ItemValue:  value,
		Dictionary: d,
	}

	if _, err := c.Client.CreateDictionaryItem(input); err != nil {
		return err
	}

	fmt.Printf("key %s and value %s successfully inserted into dictionary %s\n", key, value, dictionary)
	return nil
}

// CreateBigQueryConfig creates the Logging BigQuery configuration
func (c *Client) CreateBigQueryConfig(name, project, dataset, table, email, key, condition string) error {
	input := &fastly.CreateBigQueryInput{
		Service:           c.ServiceID,
		Version:           c.Version,
		ProjectID:         project,
		Dataset:           dataset,
		Table:             table,
		User:              email,
		ResponseCondition: condition,
		SecretKey:         key,
		Format:            "{}",
		Name:              name,
	}

	if _, err := c.Client.CreateBigQuery(input); err != nil {
		return err
	}

	fmt.Printf("bigquery configuration %s successfully created\n", name)
	return nil
}

func toSnippetType(input string) fastly.SnippetType {
	switch input {
	case "init":
		return fastly.SnippetTypeInit
	case "recv":
		return fastly.SnippetTypeRecv
	case "fetch":
		return fastly.SnippetTypeFetch
	case "hit":
		return fastly.SnippetTypeHit
	case "miss":
		return fastly.SnippetTypeMiss
	case "pass":
		return fastly.SnippetTypePass
	case "error":
		return fastly.SnippetTypeError
	case "deliver":
		return fastly.SnippetTypeDeliver
	case "log":
		return fastly.SnippetTypeLog
	case "none":
		return fastly.SnippetTypeNone
	default:
		fmt.Printf("Warning, unmatched snippet type of %s, using default snippet type of NONE", input)
		return fastly.SnippetTypeNone
	}
}
