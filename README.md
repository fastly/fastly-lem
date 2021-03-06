# fastly-lem

`fastly-lem` is a tool for helping customers automate the configuration steps necessary for setting up conditionally enabled, advanced logging to BigQuery for a Live Event Monitoring customer event.

The Live Event Monitoring logging configuration requires a specific set of steps in order for it to work correctly.  Those steps are 

1. Use our standard Advanced Logging format snippets that automate the deployment of advanced logging metrics.
2. Configure an Edge Dictionary that allows customers to conditionally turn on and off the logging since it is only needed during the event
3. Automate the setup of BigQuery as a logging provider as the steps can be a bit cumbersome to do manually 


### Configuration 
`fastly-lem` takes a parameter `-configFile` that points to the LEM configuration file, a TOML formatted config file.  Currently, the only configuration that customers need to change from the included template file (lem.config.tmpl) is the `[bigquery]` section.  The following fields are included in the configuration:

*  Top level `[fastly]` block.  The params here should not be touched or changed by customers. Inside the `[fastly]` block are 
  * `dictionary_name` the name of the Edge Dictionary used to conditionally turn on and off logging.  This value MUST be `lem_logging`
  * `logging_config_name` - The name of the bigQuery configuration in the service.  This value MUST be `LEMBigQuery`
  * `sample_rate` - A configurable string value between 0 and 100 to determine what percentage of traffic we should log. Useful if traffic volumes are very high and you only want to send a sample of traffic.
  * `log_all_errors` - A configurable string of 0 (false) or 1 (true) to determine if all errors should be logged regardless of sample_rate of logs. 0 is off and 1 is on.
  * `[[fastly.snippet]]` - An array of snippets required to make the advanced logging work
* `[bigquery]` - Customer specific BigQuery configuration   

For your own configuration, copy the template `lem.config.tmpl` to `lem.config` in a directory of your choice and edit the copy.  Remember to not modify anything underneath the `[fastly]` configuration.

### Running 

`fastly-lem` has 4 flags in order to run 

|  Parameter | Required  | Default  | Description| 
|------------|-----------|----------|------------|
| configFile | YES   | N/A  | Path to the config file that `fastly-lem` will use  |  
| token      | YES   | N/A  | Fastly API Key use to publish configuration updates   |   
| service    | YES   | N/A  | The ID of the service to use   |  
| version    | NO    | 0    | Version of the service to use.  If no version is supplied, `fastly-lem` will attempt to use the latest available `Draft` service.  If the latest version is `active`, you will need to first clone your active service in the UI.|

Example:

```fastly-lem -configFile /tmp/lem.config -token MYFASTLYTOKEN -service MYSERVICEID```

