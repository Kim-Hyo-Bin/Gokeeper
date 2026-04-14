// Command gokeeper-keygen writes a new Ed25519 PKCS#8 private key and PKIX public key PEM pair.
package main

import (
	"flag"
	"log"
	"os"

	"gokeeper/internal/signing"
)

func main() {
	privPath := flag.String("private-out", "license_private.pem", "path to write PKCS#8 private key PEM")
	pubPath := flag.String("public-out", "license_public.pem", "path to write public key PEM")
	flag.Parse()

	privPEM, pubPEM, err := signing.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*privPath, privPEM, 0o600); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*pubPath, pubPEM, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s and %s", *privPath, *pubPath)
}
