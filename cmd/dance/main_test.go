package main

import (
	"context"
	"log"
	"path/filepath"
	"testing"
)

func TestDance(t *testing.T) {
	// duplicates.yml -> duplicates
	t.Run("duplicate", func(t *testing.T) {
		matches, err := filepath.Glob("../../testdata/*.yml")
		if err != nil {
			log.Fatal(err)
		}
		run(context.Background(), matches, "ferretdb", false)
	})

}
