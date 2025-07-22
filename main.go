/*
 * Copyright (C) 2025 by John J. Rushford jrushford@apache.org
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/truenas/api_client_golang/truenas_api"
	"log"
	"os"
	"runtime/debug"
	"tnascert-deploy/config"
	"tnascert-deploy/deploy"
)

const release = "1.3"

func main() {
	// parse out command line options
	configFile := getopt.StringLong("config", 'c', config.Config_file, "full path to the configuration file")
	help := getopt.BoolLong("help", 'h', "print usage information and exit")
	version := getopt.BoolLong("version", 'v', "print version information and exit")
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
		log.Fatalln("error loading config,", err)
	}
	for i := 0; i < len(args); i++ {
		log.Printf("processing certificate installation for configuration section '%s'\n", args[i])
		cfg, ok := cfgList[args[i]]
		if !ok {
			log.Fatalf("configuration section %s not found", args[i])
		}

		serverURL := cfg.ServerURL()
		log.Printf("connecting to %s\n", cfg.ConnectHost)
		client, err := truenas_api.NewClient(serverURL, cfg.TlsSkipVerify)
		if err != nil {
			log.Println("error creating the client,", err)
			os.Exit(1)
		}
		defer func(client *truenas_api.Client) {
			err := client.Close()
			if err != nil {
				log.Printf("failed to close the client connection, %v", err)
			}
		}(client)

		// deploy the certificate key pair
		err = deploy.InstallCertificate(client, cfg)
		if err != nil {
			log.Printf("installing the certificate failed, %v", err)
		}
		log.Printf("successfully installed the certificate for configuration section '%s'\n\n", args[i])
	}
}
