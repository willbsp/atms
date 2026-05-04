package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// TODO what if not in a session?
func ListSessions() []string {
	o, _ := tmuxCmd("list-sessions", "-F", "#{session_name}")
	lines := strings.Split(strings.TrimSpace(string(o)), "\n")
	sessions := make([]string, len(lines))
	for i, l := range lines {
		l = strings.TrimSpace(l)
		sessions[i] = l
	}
	return sessions
}

func tmuxCmd(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	o, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tmux %v failed: %w", args, err)
	}

	return strings.TrimSpace(string(o)), nil
}
