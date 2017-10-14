package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/internal/legacy"
)

// Command describes a subcommand
type Command struct {
	Name  string
	Desc  string
	Flags *flag.FlagSet
	Run   func(config.Config, []string) error
}

var (
	configPath      = flag.String("config", "readeef.toml", "readeef config path")
	cpuProfile      = flag.String("cpu-profile", "", "cpu profile destination path")
	profileDuration = flag.Int("profile-duration", 5, "duration of profiling [minutes]")
	commands        = []Command{}
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *cpuProfile != "" {
		go func() {
			f, err := os.Create(*cpuProfile)
			if err != nil {
				log.Fatalf("Error creating cpu profile file %s: %v", *cpuProfile, err)
			}
			defer f.Close()

			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()

			time.Sleep(time.Duration(*profileDuration) * time.Minute)
		}()
	}

	args := flag.Args()

	if len(args) > 0 {
		for _, cmd := range commands {
			if cmd.Name == args[0] {
				cmd.Flags.Parse(args[1:])

				config, err := config.Read(*configPath)
				if err != nil {
					// Try the legacy
					config, exists, legacyErr := legacy.ReadConfig(*configPath)

					if legacyErr != nil {
						log.Fatalf("Error reading config %s: %+v", *configPath, err)
					}

					if exists {
						if err = saveLegacyConfig(config, *configPath); err != nil {
							log.Fatalf("Error saving converted legacy config %s: %+v", *configPath, err)
						}
					}
				}

				if err := cmd.Run(config, cmd.Flags.Args()); err != nil {
					log.Fatalf("Error running %s: %+v", cmd.Name, err)
				}

				os.Exit(0)
			}
		}
	}

	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s is a tool for starting and setting up
	the readeef feed aggregator.

Usage:

	readeef [flags] command [arguments]

The following flags are available:

`, os.Args[0])
	flag.PrintDefaults()

	fmt.Fprint(os.Stderr, "\nThe commands are: \n\n")

	nameLen := 0
	for _, cmd := range commands {
		if len(cmd.Name) > nameLen {
			nameLen = len(cmd.Name)
		}
	}

	for _, cmd := range commands {
		format := fmt.Sprintf("  %%%ds  %%s\n", nameLen)
		fmt.Fprintf(os.Stderr, format, cmd.Name, cmd.Desc)
	}

	fmt.Fprint(os.Stderr, "\n")
}

func saveLegacyConfig(config config.Config, path string) error {
	b, err := toml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshaling config")
	}

	file, err := os.Create(path + ".new")
	if err != nil {
		return errors.Wrapf(err, "creating new config file %s.new", path)
	}
	defer file.Close()

	if _, err = file.Write(b); err != nil {
		return errors.Wrapf(err, "writing config to %s.new", path)
	}

	if err = file.Close(); err != nil {
		return errors.Wrapf(err, "closing config file %s.new", path)
	}

	if err = os.Rename(path, path+".orig"); err != nil {
		return errors.Wrapf(err, "renameing old config file to %s.orig", path)
	}

	if err = os.Rename(path+".new", path); err != nil {
		return errors.Wrapf(err, "renameing new config file to %s", path)
	}

	return nil
}
