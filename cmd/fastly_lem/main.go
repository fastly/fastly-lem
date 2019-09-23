package main

import (
	"flag"
	"fmt"
	"github.com/fastly/fastly_lem/pkg/config"
	"os"
)

func main() {
	var (
	    configFile = flag.String("configFile","lem.config","Path to your LEM config file")
	    apiKey = flag.String("token","","API Key to use with the Fastly API")
	    serviceId = flag.String("service","","Service ID to configure")
	    version = flag.Int("version", 0, "Version of service config to use, defaults to latest")
	)

	flag.Parse()

	if *apiKey == "" {
		fmt.Printf("error, parameter 'token' is required\n")
		os.Exit(1)
	}

	if *serviceId == "" {
		fmt.Printf("error, parameter 'service' is required\n")
		os.Exit(1)
	}

	conf, err := config.New(*configFile,*apiKey, *serviceId, *version)
	if err != nil {
		fmt.Printf("error loading config file %s, %s\n",*configFile,err)
		os.Exit(1)
	}

	// Create the snippets
	if err = conf.SetupSnippets(); err != nil {
		fmt.Printf("Error creating snippets, aborting configuration: %s\n",err)
		os.Exit(1)
	}

	// Create the global condition to disable logging
	if err = conf.SetupCondition(); err != nil {
		fmt.Printf("error creating condition, aborting configuration: %s\n",err)
		os.Exit(1)
	}

	// Create the dictionary and the "enabled" key
	if err = conf.SetupDictionary(); err != nil {
		fmt.Printf("error creating dictionary, aborting configurationg: %s\n",err)
		os.Exit(1)
	}

	// Create the BigQuery configuration
	if err = conf.SetupBigQuery(); err != nil {
		fmt.Printf("error configuring BigQuery logging config: %s\n",err)
		os.Exit(1)
	}

    fmt.Printf("\n***********************************************************\n")
	fmt.Printf("Congratulations, your setup has completed successfully!\n")
	fmt.Printf("***********************************************************\n")
}