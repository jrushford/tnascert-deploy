package main

import (
	"fmt"
	"github.com/pborman/getopt/v2"
	"log"
	"os"
	"runtime/debug"
	"tnascert-deploy/clients"
	"tnascert-deploy/clients/restapi"
	"tnascert-deploy/clients/wsapi"
	"tnascert-deploy/config"
)

// application release
const release = "2.0"

func NewClient(cfg *config.Config) (clients.Client, error) {
	if cfg.ClientApi == "restapi" {
		if cfg.Debug {
			log.Printf("using a restapi client")
		}
		return restapi.NewClient(cfg)
	} else if cfg.ClientApi == "wsapi" {
		if cfg.Debug {
			log.Printf("using a wsapi client")
		}
		return wsapi.NewClient(cfg)
	}
	return nil, fmt.Errorf("empty or undefined client api in the config for %s", cfg.ConnectHost)
}

func main() {
	help := getopt.BoolLong("help", 'h', "print usage information and exit")
	version := getopt.BoolLong("version", 'v', "print version information and exit")
	configFile := getopt.StringLong("config", 'c', config.Config_file, "full path to the configuration file")
	getopt.SetParameters("config_section ... config_section")

	getopt.Parse()
	if *help == true {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}
	if *version == true {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					fmt.Printf("\nrelease: %s\ngit revision: %s\n\n", release, setting.Value)
					os.Exit(0)
				}
			}
		}
	}
	args := getopt.Args()
	if len(args) == 0 {
		args = append(args, config.Default_section)
	}

	cfgList, err := config.LoadConfig(*configFile)
	if err != nil {
		getopt.PrintUsage(os.Stdout)
		log.Fatalln("error loading the config,", err)
	}
	for i := 0; i < len(args); i++ {
		fmt.Printf("\n")
		log.Printf("processing certificate installation for '%s'\n", args[i])
		cfg, ok := cfgList[args[i]]
		if !ok {
			log.Fatalf("configuration %s was not found", args[i])
		}

		client, err := NewClient(cfg)
		if err != nil {
			log.Printf("error creating client for '%s': %v", args[i], err)
			continue
		}

		defer func(client clients.Client) {
			err := client.Close()
			if err != nil {
				log.Printf("error closing the client connection, %v", err)
			}
		}(client)

		err = client.Login()
		if err != nil {
			log.Printf("login error: %v", err)
			os.Exit(1)
		}
		err = client.PreInstall()
		if err != nil {
			log.Printf("preinstall tasks error, %v", err)
			os.Exit(1)
		}
		err = client.Install()
		if err != nil {
			log.Printf("installation tasks error, %v", err)
			os.Exit(1)
		}
		err = client.PostInstall()
		if err != nil {
			log.Printf("post installation tasks error, %v", err)
			os.Exit(1)
		}
	}
}
