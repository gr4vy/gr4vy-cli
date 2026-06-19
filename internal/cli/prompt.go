package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// isInteractive reports whether stdin is a terminal.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptLine prints a prompt to stderr and reads a trimmed line from stdin.
func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if err != nil && line == "" {
		return "", err
	}
	return line, nil
}

// promptLineDefault prompts with a default shown in brackets; an empty reply
// keeps the default.
func promptLineDefault(prompt, def string) (string, error) {
	label := prompt
	if def != "" {
		label = fmt.Sprintf("%s [%s]", prompt, def)
	}
	v, err := promptLine(label + ": ")
	if err != nil {
		return "", err
	}
	if v == "" {
		return def, nil
	}
	return v, nil
}

// promptSecret reads a single line without echoing it (for passwords).
func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// confirm asks a yes/no question, defaulting to def when the reply is empty.
func confirm(prompt string, def bool) bool {
	suffix := " [y/N]: "
	if def {
		suffix = " [Y/n]: "
	}
	v, err := promptLine(prompt + suffix)
	if err != nil {
		return def
	}
	switch strings.ToLower(v) {
	case "":
		return def
	case "y", "yes":
		return true
	default:
		return false
	}
}
