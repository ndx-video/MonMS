package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// Prompter reads interactive answers from an input stream.
type Prompter struct {
	In               io.Reader
	Out              io.Writer
	ForceInteractive bool
	reader           *bufio.Reader
}

// Stdio prompts on os.Stdin when it is a terminal.
var Stdio Prompter = Prompter{In: os.Stdin, Out: os.Stdout}

// IsInteractive reports whether prompts can be shown on this Prompter.
func (p *Prompter) IsInteractive() bool {
	if p == nil {
		return Stdio.IsInteractive()
	}
	if p.ForceInteractive {
		return true
	}
	p.ensureDefaults()
	f, ok := p.In.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}

// Confirm asks a yes/no question. Only y/yes (any case) is affirmative.
func (p *Prompter) Confirm(question string) (bool, error) {
	p.ensureDefaults()
	fmt.Fprintf(p.Out, "%s [y/N]: ", question)
	line, err := p.readLine()
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// ReadDefault prints prompt with a default value; empty input returns defaultVal.
func (p *Prompter) ReadDefault(promptText, defaultVal string) (string, error) {
	p.ensureDefaults()
	if defaultVal != "" {
		fmt.Fprintf(p.Out, "%s (%s): ", promptText, defaultVal)
	} else {
		fmt.Fprintf(p.Out, "%s: ", promptText)
	}
	line, err := p.readLine()
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}

// ReadChoice shows numbered choices; empty input selects defaultIdx.
func (p *Prompter) ReadChoice(title string, choices []string, defaultIdx int) (int, error) {
	p.ensureDefaults()
	fmt.Fprintln(p.Out, title)
	for i, c := range choices {
		fmt.Fprintf(p.Out, "  %d) %s\n", i+1, c)
	}
	def := defaultIdx + 1
	fmt.Fprintf(p.Out, "Choice [%d]: ", def)
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultIdx, nil
	}
	switch strings.ToLower(line) {
	case "1", "foreground", "fg":
		return 0, nil
	case "2", "background", "bg", "daemon":
		return 1, nil
	case "3", "exit", "quit", "q", "none":
		return 2, nil
	}
	n := 0
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(choices) {
		return defaultIdx, nil
	}
	return n - 1, nil
}

func (p *Prompter) ensureDefaults() {
	if p.In == nil {
		p.In = os.Stdin
	}
	if p.Out == nil {
		p.Out = os.Stdout
	}
	if p.reader == nil {
		p.reader = bufio.NewReader(p.In)
	}
}

func (p *Prompter) readLine() (string, error) {
	p.ensureDefaults()
	line, err := p.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}
