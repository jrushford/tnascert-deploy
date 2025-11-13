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

package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"
)

func VerifyCertificateKeyPair(cert_path string, key_path string) error {
	cert, err := tls.LoadX509KeyPair(cert_path, key_path)
	if err != nil {
		return fmt.Errorf("LoadX509KeyPair error: %v", err)
	}

	c, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("certificate parsing error: %v", err)
	}
	if time.Now().After(c.NotAfter) {
		return fmt.Errorf("Your certificate expired, a new up to date certificate is needed")
	}

	roots, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("could not load system certificate pool, %v", err)
	}
	opts := x509.VerifyOptions{
		CurrentTime: time.Now(),
		Roots:       roots,
	}
	_, err = c.Verify(opts)
	// report certificate validation information.
	if err != nil {
		log.Printf("certificate verification: %v", err)
	} else {
		log.Printf("certificate verified successfully")
	}

	return nil
}
