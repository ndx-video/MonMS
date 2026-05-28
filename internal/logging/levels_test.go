package logging_test

import (
	"testing"

	"github.com/monms/monms/internal/logging"
)

func TestDefaultLevelNames(t *testing.T) {
	t.Run("development build", func(t *testing.T) {
		logging.SetProductionBuild(false)
		got := logging.DefaultLevelNames()
		want := []string{"error", "warn", "info", "debug", "schema"}
		if len(got) != len(want) {
			t.Fatalf("got %v want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v want %v", got, want)
			}
		}
	})
	t.Run("production build", func(t *testing.T) {
		logging.SetProductionBuild(true)
		got := logging.DefaultLevelNames()
		want := []string{"error", "warn", "schema"}
		if len(got) != len(want) {
			t.Fatalf("got %v want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v want %v", got, want)
			}
		}
	})
}

func TestConfigFromValuesNilUsesDefaults(t *testing.T) {
	logging.SetProductionBuild(true)
	cfg := logging.ConfigFromValues("/tmp/site", nil, nil)
	if !cfg.Flags.Has(logging.FlagWarn) || !cfg.Flags.Has(logging.FlagSchema) {
		t.Fatalf("expected production defaults, got flags %v", cfg.Flags)
	}
	if cfg.Flags.Has(logging.FlagDebug) {
		t.Fatal("production default should not enable debug")
	}
}

func TestParseLevels(t *testing.T) {
	flags := logging.ParseLevels([]string{"warn", "schema", "debug"})
	if !flags.Has(logging.FlagError) {
		t.Fatal("error flag must always be enabled")
	}
	if !flags.Has(logging.FlagWarn) {
		t.Fatal("expected warn flag")
	}
	if !flags.Has(logging.FlagSchema) {
		t.Fatal("expected schema flag")
	}
	if !flags.Has(logging.FlagDebug) {
		t.Fatal("expected debug flag")
	}
	if flags.Has(logging.FlagInfo) {
		t.Fatal("info flag should not be set")
	}
}

func TestActiveFiles(t *testing.T) {
	cfg := logging.Config{
		Flags: logging.ParseLevels([]string{"error", "schema"}),
	}
	files := cfg.ActiveFiles()
	want := []string{"pocketbase.log", "error.log", "schema.log"}
	if len(files) != len(want) {
		t.Fatalf("got %v want %v", files, want)
	}
	for i, f := range want {
		if files[i] != f {
			t.Fatalf("files[%d]=%q want %q", i, files[i], f)
		}
	}
}
