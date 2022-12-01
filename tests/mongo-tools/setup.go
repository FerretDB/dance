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

// runCommand runs command on docker mongosh container.
func runCommand(command []string, stdout io.Writer) error {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	args := append([]string{"compose", "exec", "mongosh"}, command...)
	cmd := exec.Command(bin, args...)
	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if stdout != nil {
		cmd.Stdout = stdout
		//cmd.Stderr = stdout // TODO "dumping up to 0 collections..." is forwarded to stderr?
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(args, " "), err)
	}

	return nil
}
