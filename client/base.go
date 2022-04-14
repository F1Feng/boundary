/*
 * @Author: F1
 * @Date: 2022-03-30 16:59:40
 * @LastEditTime: 2022-03-30 16:59:41
 * @FilePath: /boundary/client/base.go
 * @Description:
 *
 */

package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/internal/cmd/base"
	"github.com/hashicorp/boundary/internal/cmd/commands/connect"
	colorable "github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
)

var conns map[string]chan struct{}

func init() {
	conns = make(map[string]chan struct{})
}

// setupEnv parses args and may replace them and sets some env vars to known
// values based on format options
func setupEnv(args []string) (retArgs []string, format string, outputCurlString bool) {
	// handle the workaround for autocomplete install/uninstall not being exported
	if len(args) == 3 &&
		args[0] == "config" &&
		args[1] == "autocomplete" {
		switch args[2] {
		case "install":
			return []string{"-autocomplete-install"}, "table", false
		case "uninstall":
			return []string{"-autocomplete-uninstall"}, "table", false
		}
	}

	var nextArgFormat bool

	for _, arg := range args {
		if nextArgFormat {
			nextArgFormat = false
			format = arg
			continue
		}

		if arg == "--" {
			break
		}

		if len(args) == 1 &&
			(arg == "-version" ||
				arg == "-v") {
			args = []string{"version"}
			break
		}

		if arg == "-output-curl-string" {
			outputCurlString = true
			continue
		}

		// Parse a given flag here, which overrides the env var
		if strings.HasPrefix(arg, "-format=") {
			format = strings.TrimPrefix(arg, "-format=")
		}
		// Handle the case where it is specified without an equal sign
		if arg == "-format" {
			nextArgFormat = true
		}
	}

	envBoundaryCLIFormat := os.Getenv(base.EnvBoundaryCLIFormat)
	// If we did not parse a value, fetch the env var
	if format == "" && envBoundaryCLIFormat != "" {
		format = envBoundaryCLIFormat
	}
	// Lowercase for consistency
	format = strings.ToLower(format)
	if format == "" {
		format = "table"
	}

	return args, format, outputCurlString
}

type RunOptions struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Address string
}

func Run(args []string) int {
	return RunCustom(args, nil)
}

// RunCustom allows passing in a base command template to pass to other
// commands. Currently, this is only used for setting a custom token helper.
func RunCustom(args []string, runOpts *RunOptions) int {
	if runOpts == nil {
		runOpts = &RunOptions{}
	}

	var format string
	var outputCurlString bool

	targetId := args[0]
	args, format, outputCurlString = setupEnv(args[1:])

	// Don't use color if disabled
	useColor := true
	if os.Getenv(base.EnvBoundaryCLINoColor) != "" || color.NoColor {
		useColor = false
	}

	if runOpts.Stdout == nil {
		runOpts.Stdout = os.Stdout
	}
	if runOpts.Stderr == nil {
		runOpts.Stderr = os.Stderr
	}

	// Only use colored UI if stdout is a tty, and not disabled
	if useColor && format == "table" {
		if f, ok := runOpts.Stdout.(*os.File); ok {
			runOpts.Stdout = colorable.NewColorable(f)
		}
		if f, ok := runOpts.Stderr.(*os.File); ok {
			runOpts.Stderr = colorable.NewColorable(f)
		}
	} else {
		runOpts.Stdout = colorable.NewNonColorable(runOpts.Stdout)
		runOpts.Stderr = colorable.NewNonColorable(runOpts.Stderr)
	}

	uiErrWriter := runOpts.Stderr
	if outputCurlString {
		uiErrWriter = ioutil.Discard
	}

	ui := &base.BoundaryUI{
		Ui: &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			Ui: &cli.BasicUi{
				Reader:      bufio.NewReader(os.Stdin),
				Writer:      runOpts.Stdout,
				ErrorWriter: uiErrWriter,
			},
		},
		Format: format,
	}

	switch format {
	case "table", "json":
	default:
		ui.Error(fmt.Sprintf("Invalid output format: %s", format))
		return 1
	}

	hiddenCommands := []string{"version"}

	cmd := NewCommand(ui)

	conns[targetId] = cmd.ShutdownCh

	cli := &cli.CLI{
		Name: "boundary",
		Args: args,
		Commands: map[string]cli.CommandFactory{
			"connect": func() (cli.Command, error) {
				return &connect.Command{
					Command: cmd,
					Func:    "connect",
				}, nil
			},
		},
		HelpFunc: groupedHelpFunc(
			cli.BasicHelpFunc("boundary"),
		),
		HelpWriter:                 runOpts.Stderr,
		HiddenCommands:             hiddenCommands,
		Autocomplete:               true,
		AutocompleteNoDefaultFlags: true,
	}

	go func() {
		exitCode, err := cli.Run()
		if outputCurlString {
			if exitCode == 0 {
				fmt.Fprint(runOpts.Stderr, "Could not generate cURL command\n")
				return
			} else {
				if api.LastOutputStringError == nil {
					if exitCode == 127 {
						// Usage, just pass it through
						return
					}
					fmt.Fprint(runOpts.Stderr, "cURL command not set by API operation; run without -output-curl-string to see the generated error\n")
					return
				}
				if !strings.Contains(api.LastOutputStringError.Error(), api.ErrOutputStringRequest) {
					runOpts.Stdout.Write([]byte(fmt.Sprintf("Error creating request string: %s\n", api.LastOutputStringError.Error())))
					return
				}
				runOpts.Stdout.Write([]byte(fmt.Sprintf("%s\n", api.LastOutputStringError.CurlString())))
				return
			}
		} else if err != nil {
			fmt.Fprintf(runOpts.Stderr, "Error executing CLI: %s\n", err.Error())
			return
		}
	}()

	return 0
}

func groupedHelpFunc(f cli.HelpFunc) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var b bytes.Buffer
		tw := tabwriter.NewWriter(&b, 0, 2, 6, ' ', 0)

		fmt.Fprintf(tw, "Usage: boundary <command> [args]\n")

		otherCommands := make([]string, 0, len(commands))
		for k := range commands {
			otherCommands = append(otherCommands, k)
		}
		sort.Strings(otherCommands)

		fmt.Fprintf(tw, "\n")
		fmt.Fprintf(tw, "Commands:\n")
		for _, v := range otherCommands {
			printCommand(tw, v, commands[v])
		}

		tw.Flush()

		return strings.TrimSpace(b.String())
	}
}

func printCommand(w io.Writer, name string, cmdFn cli.CommandFactory) {
	cmd, err := cmdFn()
	if err != nil {
		panic(fmt.Sprintf("failed to load %q command: %s", name, err))
	}
	fmt.Fprintf(w, "    %s\t%s\n", name, cmd.Synopsis())
}

func disconnect(targetId string) error {
	if _, ok := conns[targetId]; ok {
		conns[targetId] <- struct{}{}
		delete(conns, targetId)
	} else {
		return fmt.Errorf("failed disconnect by targetId: %s, not exists", targetId)
	}
	return nil
}

// New returns a new instance of a base.Command type
func NewCommand(ui cli.Ui) *base.Command {
	ctx, cancel := context.WithCancel(context.Background())
	ret := &base.Command{
		UI:         ui,
		ShutdownCh: make(chan struct{}),
		Context:    ctx,
	}

	go func() {
		<-ret.ShutdownCh

		println("shutdown...")
		cancel()
	}()

	return ret
}
