package mongotools

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func setupMongoContainer(t *testing.T) {
	t.Helper()
}

// runCommand runs command with given arguments.
func runCommand(command string, args []string, stdout io.Writer) error {
	bin, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = os.Stdout
	if stdout != nil {
		cmd.Stdout = stdout
	}
	cmd.Stderr = os.Stderr

	//cmd.Stderr = bytes.NewBuffer([]byte{})

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(args, " "), err)
	}

	return nil
}
