package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ListSessions() []string {
	o, _ := tmuxCmd("list-sessions", "-F", "#{session_name}")
	if o == "" {
		return []string{}
	}
	lines := strings.Split(o, "\n")
	sessions := make([]string, len(lines))
	for i, l := range lines {
		sessions[i] = strings.TrimSpace(l)
	}
	return sessions
}

func SwitchOrAttach(path string) error {
	name := SessionName(path)
	if !HasSession(name) {
		if o, err := tmuxCmd("new-session", "-d", "-s", name, "-c", path); err != nil {
			return fmt.Errorf("failed to create session %q: %s (%w)", name, o, err)
		}
	}
	if IsInsideTmux() {
		_, err := tmuxCmd("switch-client", "-t", name)
		return err
	}

	_, err := tmuxCmd("attach-session", "-t", name)
	return err
}

func SessionName(repoPath string) string {
	name := filepath.Base(repoPath)
	return name
}

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func HasSession(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

func tmuxCmd(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	o, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tmux %v failed: %w", args, err)
	}

	return strings.TrimSpace(string(o)), nil
}
