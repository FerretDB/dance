package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func Test(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		filePath := filepath.Join("..", "..", "testdata", fmt.Sprint(t.Name, ".yml"))

	})
}
