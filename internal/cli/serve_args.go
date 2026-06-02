package cli

// EnsureServeSubcommand prepends "serve" when args begin with flags or the "start" synonym.
// MonMS defaults to serve when no subcommand is given (monms, monms -d, monms --dev).
func EnsureServeSubcommand(args []string) []string {
	args = stripStartSynonym(args)
	if len(args) == 0 {
		return []string{"serve"}
	}
	switch args[0] {
	case "serve", "init", "validate", "content", "stop", "restart", "superuser", "migrate":
		return args
	default:
		return append([]string{"serve"}, args...)
	}
}

// stripStartSynonym removes a leading "start" so PocketBase does not treat it as a domain.
func stripStartSynonym(args []string) []string {
	if len(args) > 0 && args[0] == "start" {
		return args[1:]
	}
	return args
}
