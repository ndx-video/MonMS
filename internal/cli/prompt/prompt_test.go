package prompt

import (
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	t.Parallel()

	yes := &Prompter{In: strings.NewReader("yes\n"), Out: ioDiscard{}}
	ok, err := yes.Confirm("create?")
	if err != nil || !ok {
		t.Fatalf("Confirm(yes) = %v, %v; want true, nil", ok, err)
	}

	no := &Prompter{In: strings.NewReader("\n"), Out: ioDiscard{}}
	ok, err = no.Confirm("create?")
	if err != nil || ok {
		t.Fatalf("Confirm(empty) = %v, %v; want false, nil", ok, err)
	}
}

func TestReadDefault(t *testing.T) {
	t.Parallel()

	p := &Prompter{In: strings.NewReader("\n"), Out: ioDiscard{}}
	got, err := p.ReadDefault("port", "8090")
	if err != nil || got != "8090" {
		t.Fatalf("ReadDefault() = %q, %v", got, err)
	}

	p2 := &Prompter{In: strings.NewReader("9000\n"), Out: ioDiscard{}}
	got, err = p2.ReadDefault("port", "8090")
	if err != nil || got != "9000" {
		t.Fatalf("ReadDefault(custom) = %q, %v", got, err)
	}
}

func TestReadChoice(t *testing.T) {
	t.Parallel()
	choices := []string{"foreground", "background", "exit without starting"}

	p := &Prompter{In: strings.NewReader("2\n"), Out: ioDiscard{}}
	idx, err := p.ReadChoice("start?", choices, 0)
	if err != nil || idx != 1 {
		t.Fatalf("ReadChoice(2) = %d, %v; want 1, nil", idx, err)
	}

	p2 := &Prompter{In: strings.NewReader("\n"), Out: ioDiscard{}}
	idx, err = p2.ReadChoice("start?", choices, 0)
	if err != nil || idx != 0 {
		t.Fatalf("ReadChoice(default) = %d, %v; want 0, nil", idx, err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
