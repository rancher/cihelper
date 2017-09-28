package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func Clone(path, url, branch string) error {
	log.Infof("args:%v,%v,%v", branch, url, path)
	return runcmd("git", "clone", "-b", branch, "--single-branch", url, path)
}

func Update(path, branch string) error {
	if err := runcmd("git", "-C", path, "fetch"); err != nil {
		return err
	}
	return runcmd("git", "-C", path, "checkout", fmt.Sprintf("origin/%s", branch))
}

//LazyPush add,commit and push changes.
func LazyPush(path, repo, refspec string) error {

	if err := runcmd("git", "-C", path, "add", "."); err != nil {
		return err
	}
	if err := runcmd("git", "-C", path, "commit", "-m", "update"); err != nil {
		return err
	}
	return runcmd("git", "-C", path, "push", repo, refspec)
}

func HeadCommit(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
	output, err := cmd.Output()
	return strings.Trim(string(output), "\n"), err
}

func IsValid(url string) bool {
	err := runcmd("git", "ls-remote", url)
	return (err == nil)
}

func runcmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
