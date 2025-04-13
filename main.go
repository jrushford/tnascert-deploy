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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pborman/getopt/v2"
	"log"
	"os"
	"tnascert-deploy/config"
	"tnascert-deploy/deploy"
)

// simple verification of the certificate and private key, can they be loaded and parsed
func verifyCertificateKeyPair(cert_path string, key_path string) error {
	cert, err := tls.LoadX509KeyPair(cert_path, key_path)
	if err != nil {
		return fmt.Errorf("LoadX509KeyPair error: %v", err)
	}
	_, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("ParseCertificate error: %v", err)
	}
	return nil
}

func main() {
	var section string = config.Default_section

	// parse out command line options
	config_file := getopt.StringLong("config", 'c', config.Config_file, "full path to the configuration file")
	help := getopt.BoolLong("help", 'h', "print usage information and exit")
	getopt.SetParameters("ini_section_name")

	getopt.Parse()
	if *help == true {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}
	args := getopt.Args()
	if len(args) > 0 {
		section = args[0]
	}

	cfg, err := config.New(*config_file, section)
	if err != nil {
		log.Fatalln("Error loading config,", err)
	}

	// run a simple check of the certificate and private key before deployment.
	err = verifyCertificateKeyPair(cfg.Fullchain_path, cfg.Private_key_path)
	if err != nil {
		log.Fatalf("verifying certificate key pair, %v", err)
	} else {
		log.Println("verified certificate key pair")
	}

	// deploy the certificate key pair
	err = deploy.InstallCertificate(cfg)
	if err != nil {
		log.Printf("Installing certificate, %v", err)
	}
}
