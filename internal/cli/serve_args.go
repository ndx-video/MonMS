package cli

// EnsureServeSubcommand prepends "serve" when args begin with flags.
// MonMS defaults to serve when no subcommand is given (monms, monms -d, monms --dev).
func EnsureServeSubcommand(args []string) []string {
	if len(args) == 0 {
		return []string{"serve"}
	}
	switch args[0] {
	case "serve", "init", "validate", "content", "stop", "superuser", "migrate":
		return args
	default:
		return append([]string{"serve"}, args...)
	}
}
