package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type SessionTarget interface {
	Name() string
	Dir() string
}

type SessionSet map[string]struct{}

func (s SessionSet) Has(t SessionTarget) bool {
	_, ok := s[t.Name()]
	return ok
}

func LoadSessions() SessionSet {
	names := listSessions()
	set := make(SessionSet, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		set[n] = struct{}{}
	}
	return set
}

func SwitchOrAttach(t SessionTarget) error {
	name := t.Name()
	if !hasSession(name) {
		if o, err := tmuxCmd("new-session", "-d", "-s", name, "-c", t.Dir()); err != nil {
			return fmt.Errorf("failed to create session %q: %s (%w)", name, o, err)
		}
	}
	if isInsideTmux() {
		_, err := tmuxCmd("switch-client", "-t", name)
		return err
	}

	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}

	return syscall.Exec(tmux, []string{"tmux", "attach-session", "-t", name}, os.Environ())
}

func listSessions() []string {
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

func isInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func hasSession(name string) bool {
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
