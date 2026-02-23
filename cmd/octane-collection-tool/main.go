package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/octane-collection-tool/internal/config"
	"github.com/example/octane-collection-tool/internal/junit"
	"github.com/example/octane-collection-tool/internal/octane"
)

type typedValues []junit.TypeValue

func (t *typedValues) String() string {
	parts := make([]string, 0, len(*t))
	for _, v := range *t {
		parts = append(parts, v.Type+":"+v.Value)
	}
	return strings.Join(parts, ",")
}

func (t *typedValues) Set(value string) error {
	entry, err := parseTypeValue(value)
	if err != nil {
		return err
	}
	*t = append(*t, entry)
	return nil
}

func parseTypeValue(value string) (junit.TypeValue, error) {
	parts := strings.Split(value, ":")
	if len(parts) < 2 || parts[0] == "" {
		return junit.TypeValue{}, fmt.Errorf("invalid TYPE:VALUE entry: %s", value)
	}
	return junit.TypeValue{Type: parts[0], Value: strings.Join(parts[1:], ":")}, nil
}

func main() {
	var cfgFile, server, sharedSpace, workspace, user, password, accessToken string
	var release, suite, outputFile, startedRaw string
	var releaseDefault, internal, checkResult bool
	var checkResultTimeout int
	var tags, fields typedValues

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&cfgFile, "config-file", "", "configuration file location")
	fs.StringVar(&server, "server", "", "server URL")
	fs.StringVar(&sharedSpace, "shared-space", "", "shared space")
	fs.StringVar(&workspace, "workspace", "", "workspace")
	fs.StringVar(&user, "user", "", "username")
	fs.StringVar(&password, "password", "", "password")
	fs.StringVar(&accessToken, "access-token", "", "bearer token")
	fs.Var(&tags, "tag", "environment tag TYPE:VALUE")
	fs.Var(&fields, "field", "field tag TYPE:VALUE")
	fs.StringVar(&release, "release", "", "release ID")
	fs.BoolVar(&releaseDefault, "release-default", false, "use _default_ release")
	fs.StringVar(&suite, "suite", "", "suite ID")
	fs.StringVar(&startedRaw, "started", "", "start timestamp milliseconds")
	fs.BoolVar(&internal, "internal", false, "files are internal XML")
	fs.StringVar(&outputFile, "output-file", "", "write internal XML")
	fs.BoolVar(&checkResult, "check-result", false, "poll result status")
	fs.IntVar(&checkResultTimeout, "check-result-timeout", 10, "poll timeout seconds")

	fs.StringVar(&cfgFile, "c", "", "configuration file location")
	fs.StringVar(&server, "s", "", "server URL")
	fs.StringVar(&sharedSpace, "d", "", "shared space")
	fs.StringVar(&workspace, "w", "", "workspace")
	fs.StringVar(&user, "u", "", "username")
	fs.StringVar(&password, "p", "", "password")
	fs.Var(&tags, "t", "environment tag TYPE:VALUE")
	fs.Var(&fields, "f", "field tag TYPE:VALUE")
	fs.StringVar(&release, "r", "", "release ID")
	fs.BoolVar(&internal, "i", false, "files are internal XML")
	fs.StringVar(&outputFile, "o", "", "write internal XML")

	if err := fs.Parse(os.Args[1:]); err != nil {
		exitErr(err)
	}
	files := fs.Args()
	if len(files) == 0 {
		exitErr(errors.New("at least one input file is required"))
	}
	if outputFile != "" && len(files) != 1 {
		exitErr(errors.New("output mode supports exactly one input JUnit file"))
	}
	if !internal && suite != "" && release == "" && !releaseDefault {
		exitErr(errors.New("suite injection requires --release <id> or --release-default"))
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		exitErr(err)
	}
	effective := octane.Config{
		Server:      choose(server, cfg["server"]),
		SharedSpace: choose(sharedSpace, cfg["sharedspace"]),
		Workspace:   choose(workspace, cfg["workspace"]),
		User:        choose(user, cfg["user"]),
		Password:    choose(password, cfg["password"]),
		AccessToken: accessToken,
	}

	started := time.Now().UnixMilli()
	if startedRaw != "" {
		if parsed, err := strconv.ParseInt(startedRaw, 10, 64); err == nil {
			started = parsed
		}
	}

	client := octane.NewClient(effective)
	for _, file := range files {
		var payload string
		if internal {
			bytes, err := os.ReadFile(file)
			if err != nil {
				exitErr(err)
			}
			payload = string(bytes)
		} else {
			xml, err := junit.BuildInternalResultXML(junit.Options{
				InputPath:      file,
				Started:        started,
				Tags:           tags,
				Fields:         fields,
				Release:        release,
				ReleaseDefault: releaseDefault,
				Suite:          suite,
			})
			if err != nil {
				exitErr(err)
			}
			payload = xml
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(payload), 0644); err != nil {
				exitErr(err)
			}
			fmt.Printf("Internal XML was written to %s\n", outputFile)
			continue
		}

		if effective.Server == "" || effective.SharedSpace == "" || effective.Workspace == "" {
			exitErr(errors.New("missing required server configuration (server/shared-space/workspace)"))
		}

		resp, err := client.PushResult(payload)
		if err != nil {
			exitErr(err)
		}

		fmt.Printf("Pushed test result file: %s", file)
		if resp.ID != "" {
			status := resp.Status
			if status == "" {
				status = "unknown"
			}
			fmt.Printf(" (id=%s, status=%s)", resp.ID, status)
		}
		fmt.Println()

		if checkResult && resp.ID != "" {
			final, err := client.WaitForResultCompletion(resp.ID, checkResultTimeout, 100*time.Millisecond)
			if err != nil {
				exitErr(err)
			}
			fmt.Printf("Result %s final status: %s", resp.ID, choose(final.Status, "unknown"))
			if final.TimedOut {
				fmt.Printf(" (timeout reached)")
			}
			if final.ErrorDetails != "" {
				fmt.Printf(", msg: %s", final.ErrorDetails)
			}
			fmt.Println()
		}
	}
}

func choose(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
