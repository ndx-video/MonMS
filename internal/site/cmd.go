package site

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/config"
)

// RunCLI is the entry point for `monms site` subcommands.
func RunCLI(args []string) error {
	if len(args) == 0 {
		cli.PrintHelp("site")
		return nil
	}

	switch args[0] {
	case "sync":
		return runSyncCLI(args[1:])
	default:
		return fmt.Errorf("unknown site subcommand %q (want sync)", args[0])
	}
}

func runSyncCLI(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		if text, ok := cli.SiteSubcommandHelp("sync"); ok {
			fmt.Print(text)
		}
	}

	var ref, remote string
	var force bool
	fs.StringVar(&ref, "ref", "", "git ref to checkout (tag or branch, required)")
	fs.StringVar(&remote, "remote", "origin", "remote name to fetch from")
	fs.BoolVar(&force, "force", false, "checkout even when the worktree has local changes")

	if err := fs.Parse(config.StripSiteFlags(args)); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
			return nil
		}
		return err
	}

	_, siteAbs, err := config.ResolveSite(args, os.Environ())
	if err != nil {
		return err
	}

	if ref == "" {
		return fmt.Errorf("site sync: --ref is required")
	}

	return Sync(siteAbs, SyncOptions{
		Ref:    ref,
		Remote: remote,
		Force:  force,
	})
}
