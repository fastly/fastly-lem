package api

import (
	"errors"
	"fmt"
	"github.com/fastly/go-fastly/fastly"
)

type ApiClient struct {
	Client *fastly.Client
	ServiceId string
	Version int
}

// New - create a new instance of ApiClient.
// Automatically set the version to the latest version we can
// write to unless it was supplied
func New(key string, id string, version int) (ApiClient, error) {
	client, err := fastly.NewClient(key)
	if err != nil {
		return ApiClient{}, err
	}

	if version == 0 {
		latest, err := GetLatestVersion(key,id)
		if err != nil {
			return ApiClient{},err
		}
		fmt.Printf("Using latest writeable version %d, since no version was provided\n",latest)
		version = latest
	}

	c := ApiClient{
		client,
		id,
		version,
	}

	return c, nil
}

// GetLatestVersion - figures out the latest version of the configuration
func GetLatestVersion(key string, serviceId string) (int,error) {
	client, err := fastly.NewClient(key)
	if err != nil {
		return 0, err
	}
	latest, err := client.LatestVersion(&fastly.LatestVersionInput{
		Service: serviceId,
	})
	if err != nil {
		return -1, err
	}

	// API returns nil if there is no version of the config
	if latest == nil {
		return -1, errors.New("there is currently no version in the service, please create a version in your service first")
	}

	// Check if latest version is not active
	if latest.Active == true {
		// TBD see if we want to automatically clone for user?
		//newVersion, err := client.CloneVersion(&fastly.CloneVersionInput{
		//	Service: serviceId,
		//	Version: latest.Number,
		//})
		// fmt.Printf("cloning active version %d to new version %d since there is no draft config to write to",latest.Number, newVersion)
		msg := fmt.Sprintf("there is no writeable config version, the latest version %d is active and should be cloned before continuing",latest.Number)
		return -1, errors.New(msg)
	}

	return latest.Number, nil
}

// CreateSnippet - creates a snippet in the service configuration
func (c *ApiClient) CreateSnippet(name string, content string, priority int, snippetType string) error {
	input := &fastly.CreateSnippetInput{
		Service: c.ServiceId,
		Version: c.Version,
		Name: name,
		Priority: priority,
		Type: toSnippetType(snippetType),
		Content: content,

	}

	_, err := c.Client.CreateSnippet(input)
	if err == nil {
		fmt.Printf("Snippet %s created for method %s in version %d\n",name,snippetType,c.Version)
	}

	return err
}

// SetupCondition - create a new condition in the API that we can attach to other objects
func (c *ApiClient) CreateCondition(name string, statement string, priority int, condType string) error {
	input := &fastly.CreateConditionInput{
		Service: c.ServiceId,
		Version: c.Version,
		Name: name,
		Statement: statement,
		Priority: priority,
		Type: condType,
	}

	_, err := c.Client.CreateCondition(input)
	if err == nil {
		fmt.Printf("Condition %s created with type %s in version %d\n",name,condType,c.Version)
	}

	return err
}

// CreateDictionary - Creates a new edge dictionary
func (c *ApiClient) CreateDictionary(name string) error {
	input := &fastly.CreateDictionaryInput{
		Service: c.ServiceId,
		Version: c.Version,
		Name: name,
	}

	_, err := c.Client.CreateDictionary(input)
	if err == nil {
		fmt.Printf("Dictionary %s successfully created\n",name)
	}

	return err
}

// CreateDictionaryItem - Creates a new element in the dictionary
func (c *ApiClient) CreateDictionaryItem(dictionary string, key string, value string) error {
	input := &fastly.CreateDictionaryItemInput{
		Service: c.ServiceId,
		ItemKey: key,
		ItemValue: value,
		Dictionary: dictionary,
	}

	_, err := c.Client.CreateDictionaryItem(input)
	if err == nil {
		fmt.Printf("key %s and value %s successfully inserted into dictionary %s\n",key,value,dictionary)
	}

	return err
}

// CreateBigQueryConfig - Creates the Logging BigQuery configuration
func (c *ApiClient) CreateBigQueryConfig(name, project, dataset, table, email, key, condition string, ) error {
	input := &fastly.CreateBigQueryInput{
	  Service: c.ServiceId,
	  Version: c.Version,
	  ProjectID: project,
	  Dataset: dataset,
	  Table: table,
	  User: email,
	  ResponseCondition: condition,
	  SecretKey: key,
	  Format: "{}",
	}

	_, err := c.Client.CreateBigQuery(input)
	if err == nil {
		fmt.Printf("bigquery configuration %s successfully created\n",name)
	}

	return err
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
		fmt.Printf("Warning, unmatched snippet type of %s, using default snippet type of NONE",input)
		return fastly.SnippetTypeNone
	}
}