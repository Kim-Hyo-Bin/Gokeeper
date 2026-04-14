// HTTP API examples: issue (7-day), get, verify, revoke.
//
// Run from repo: cd examples/go/api && go run .
//
// Env: BASE_URL (default http://localhost:8080).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	base := os.Getenv("BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	c := &http.Client{Timeout: 15 * time.Second}

	exp := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339Nano)
	body, _ := json.Marshal(map[string]string{"expires_at": exp})
	resp, err := c.Post(base+"/v1/licenses", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var issue struct {
		ID         string `json:"id"`
		LicenseKey string `json:"license_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("issue: %s", resp.Status)
	}
	fmt.Println("issued id:", issue.ID)

	resp2, err := c.Get(base + "/v1/licenses/" + issue.ID)
	if err != nil {
		log.Fatal(err)
	}
	resp2.Body.Close()
	fmt.Println("get status:", resp2.Status)

	vbody, _ := json.Marshal(map[string]string{"license_key": issue.LicenseKey})
	resp3, err := c.Post(base+"/v1/licenses/verify", "application/json", bytes.NewReader(vbody))
	if err != nil {
		log.Fatal(err)
	}
	var v map[string]any
	json.NewDecoder(resp3.Body).Decode(&v)
	resp3.Body.Close()
	fmt.Println("verify:", v)

	req, _ := http.NewRequest(http.MethodPost, base+"/v1/licenses/"+issue.ID+"/revoke", nil)
	resp4, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp4.Body.Close()
	fmt.Println("revoke:", resp4.Status)
}
