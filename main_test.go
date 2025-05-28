package main

import (
	"fmt"
	"testing"
	"tnascert-deploy/config"
)

func TestConfig(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"

	cfg, err := config.New(configFile, "default")
	if cfg == nil || err != nil {
		t.Fatalf("New config failed with error: %v", err)
	}

	err = verifyCertificateKeyPair(cfg.FullChainPath, cfg.Private_key_path)
	if err != nil {
		t.Fatalf("verifying the certificate key pair, %v", err)
	} else {
		fmt.Println("verified the certificate key pair")
	}
}
