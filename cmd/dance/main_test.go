package main

import (
	"context"
	"log"
	"path/filepath"
	"testing"
)

// Runs tests from testdata directory.
func TestDance(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		matches, err := filepath.Glob("../../testdata/*.yml")
		if err != nil {
			log.Fatal(err)
		}
		run(context.Background(), matches, "ferretdb", true)
	})
}
