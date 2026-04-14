// Offline license validation using the same logic as the server-issued wire format.
//
// Run from repo: cd examples/go/offline && go run .
//
// Env: PUBLIC_KEY_PATH (default ./license_public.pem), LICENSE_KEY (required).
package main

import (
	"fmt"
	"log"
	"os"

	"gokeeper/pkg/license"
)

func main() {
	pemPath := os.Getenv("PUBLIC_KEY_PATH")
	if pemPath == "" {
		pemPath = "license_public.pem"
	}
	keyStr := os.Getenv("LICENSE_KEY")
	if keyStr == "" {
		log.Fatal("set LICENSE_KEY to the issued license_key string")
	}
	pemBytes, err := os.ReadFile(pemPath)
	if err != nil {
		log.Fatal(err)
	}
	claims, err := license.Validate(keyStr, pemBytes)
	if err != nil {
		log.Fatalf("invalid: %v", err)
	}
	fmt.Printf("ok license_id=%s exp=%d\n", claims.LicenseID, claims.Exp)
}
