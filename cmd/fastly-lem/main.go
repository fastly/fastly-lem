package main

import (
	"flag"
	"fmt"
	"github.com/fastly/fastly-lem/pkg/config"
	"os"
)

const CLIVersion = "1.1"

func main() {
	var (
		configFile = flag.String("configFile", "lem.config", "Path to your LEM config file")
		apiKey     = flag.String("token", "", "API Key to use with the Fastly API")
		serviceID  = flag.String("service", "", "Service ID to configure")
		version    = flag.Int("version", 0, "Version of service config to use, defaults to latest")
	)

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "fastly-lem:  A CLI tool for configuring Live Event Monitoring\n")
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\n\n", CLIVersion)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *apiKey == "" {
		fmt.Printf("error, parameter 'token' is required\n")
		os.Exit(1)
	}

	if *serviceID == "" {
		fmt.Printf("error, parameter 'service' is required\n")
		os.Exit(1)
	}

	conf, err := config.New(*configFile, *apiKey, *serviceID, *version)
	if err != nil {
		fmt.Printf("error loading config file %s: %v\n", *configFile, err)
		os.Exit(1)
	}

	// Create the snippets
	if err = conf.SetupSnippets(); err != nil {
		fmt.Printf("error creating snippets, aborting configuration: %v\n", err)
		os.Exit(1)
	}

	// Create the dictionary and the keys
	if err = conf.SetupDictionary(); err != nil {
		fmt.Printf("error creating dictionary, aborting configuration: %v\n", err)
		os.Exit(1)
	}

	// Create the BigQuery configuration
	if err = conf.SetupBigQuery(); err != nil {
		fmt.Printf("error configuring BigQuery logging config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n***********************************************************\n")
	fmt.Printf("Congratulations, your setup has completed successfully!\n")
	fmt.Printf("***********************************************************\n")
}
