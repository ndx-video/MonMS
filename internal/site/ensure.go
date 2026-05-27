package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/cli/prompt"
	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/scaffold"
)

// ErrDeclined is returned when the operator declines interactive scaffolding.
var ErrDeclined = errors.New("site setup declined")

// EnsureOutcome is the result of pre-serve site preparation.
type EnsureOutcome struct {
	SiteAbs   string
	StartMode scaffold.StartMode
	Scaffolded bool
}

// EnsureReady validates the site and interactively scaffolds when needed.
func EnsureReady(res config.SiteResolution, p *prompt.Prompter) (EnsureOutcome, error) {
	siteAbs := res.Absolute
	status, missing := CheckSite(siteAbs)
	if status == StatusReady {
		return EnsureOutcome{SiteAbs: siteAbs, StartMode: scaffold.StartUnset}, nil
	}

	if !p.IsInteractive() {
		return EnsureOutcome{}, nonInteractiveError(res, status, missing)
	}

	ok, err := confirmScaffold(p, res, status)
	if err != nil {
		return EnsureOutcome{}, err
	}
	if !ok {
		return EnsureOutcome{}, ErrDeclined
	}

	if status == StatusMissingRoot && !res.SiteFlagSet {
		path, err := p.ReadDefault("Site path", "site")
		if err != nil {
			return EnsureOutcome{}, err
		}
		path = filepath.Clean(strings.TrimSpace(path))
		if path == "" || path == "." {
			return EnsureOutcome{}, fmt.Errorf("site path must not be empty")
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return EnsureOutcome{}, fmt.Errorf("resolve site absolute path: %w", err)
		}
		siteAbs = abs
	}

	result, err := scaffold.InitAt(siteAbs)
	if err != nil {
		return EnsureOutcome{}, err
	}
	printCreated(result)

	mode, err := scaffold.RunSetupWizard(siteAbs, p)
	if err != nil {
		return EnsureOutcome{}, err
	}

	if st, _ := CheckSite(siteAbs); st != StatusReady {
		return EnsureOutcome{}, fmt.Errorf("site still incomplete after init")
	}

	return EnsureOutcome{
		SiteAbs:    siteAbs,
		StartMode:  mode,
		Scaffolded: true,
	}, nil
}

func confirmScaffold(p *prompt.Prompter, res config.SiteResolution, status Status) (bool, error) {
	switch {
	case status == StatusMissingRoot && !res.SiteFlagSet && res.Configured == "./site":
		return p.Confirm("The default './site' folder was not found, and you did not specify (-s|--site) an alternative folder. Would you like me to create and initialize one?")
	case status == StatusMissingRoot:
		return p.Confirm(fmt.Sprintf("Site folder %s was not found. Would you like me to create and initialize it?", res.Configured))
	default:
		return p.Confirm(fmt.Sprintf("Folder %s exists, but does not contain the required files. Would you like me to add them to the folder you specified?", res.Configured))
	}
}

func nonInteractiveError(res config.SiteResolution, status Status, missing []string) error {
	switch status {
	case StatusMissingRoot:
		if !res.SiteFlagSet && res.Configured == "./site" {
			return fmt.Errorf("site not found at ./site (default); run monms init or monms serve -s PATH")
		}
		return fmt.Errorf("site not found at %s; run monms init -s %q", res.Configured, res.Configured)
	default:
		if len(missing) > 0 {
			return fmt.Errorf("site incomplete at %s: missing %s; run monms init -s %q", res.Configured, missing[0], res.Configured)
		}
		return fmt.Errorf("site incomplete at %s; run monms init -s %q", res.Configured, res.Configured)
	}
}

func printCreated(result *scaffold.Result) {
	if result == nil || len(result.Created) == 0 {
		return
	}
	fmt.Fprintln(os.Stdout, "Created:")
	for _, p := range result.Created {
		fmt.Println(p)
	}
}
