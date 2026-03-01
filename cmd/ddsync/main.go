package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/metalagman/ddgo/internal/ddsync"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	cmd := args[0]
	cfg, err := parseConfig(args[1:])
	if err != nil {
		return err
	}

	switch cmd {
	case "update":
		manifest, err := ddsync.Update(cfg)
		if err != nil {
			return err
		}
		fmt.Printf("updated snapshot %s with %d files\n", manifest.SnapshotVersion, len(manifest.SourceFiles))
		return nil
	case "verify":
		if err := ddsync.Verify(cfg); err != nil {
			return err
		}
		fmt.Println("verify: clean")
		return nil
	case "status":
		report, err := ddsync.Status(cfg)
		if err != nil {
			return err
		}
		if report.Clean {
			fmt.Println("status: clean")
			return nil
		}
		fmt.Println("status: dirty")
		for _, issue := range report.Issues {
			fmt.Println("-", issue)
		}
		return nil
	case "help", "-h", "--help":
		return usageError()
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func parseConfig(args []string) (ddsync.Config, error) {
	fs := flag.NewFlagSet("ddsync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	snapshotDir := fs.String("snapshot-dir", "sync/snapshots/v1", "Path to pinned snapshot directory")
	outputPath := fs.String("output", "sync/compiled.json", "Path to generated deterministic artifact")
	manifestPath := fs.String("manifest", "sync/manifest.json", "Path to generated manifest")
	version := fs.String("version", "v1", "Pinned snapshot version")

	if err := fs.Parse(args); err != nil {
		return ddsync.Config{}, err
	}

	return ddsync.Config{
		Version:      strings.TrimSpace(*version),
		SnapshotDir:  strings.TrimSpace(*snapshotDir),
		OutputPath:   strings.TrimSpace(*outputPath),
		ManifestPath: strings.TrimSpace(*manifestPath),
	}, nil
}

func usageError() error {
	return errors.New("usage: ddsync <update|verify|status> [--snapshot-dir DIR --output FILE --manifest FILE --version V]")
}
