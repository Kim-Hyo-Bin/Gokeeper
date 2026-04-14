package model

import "testing"

func TestLicense_TableName(t *testing.T) {
	var l License
	if l.TableName() != "licenses" {
		t.Fatal(l.TableName())
	}
}
