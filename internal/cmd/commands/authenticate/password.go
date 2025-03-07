package authenticate

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/internal/cmd/base"
	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-wordwrap"
	"github.com/posener/complete"
)

var (
	_ cli.Command             = (*PasswordCommand)(nil)
	_ cli.CommandAutocomplete = (*PasswordCommand)(nil)
)

var (
	envPassword  = "BOUNDARY_AUTHENTICATE_PASSWORD_PASSWORD"
	envLoginName = "BOUNDARY_AUTHENTICATE_PASSWORD_LOGIN_NAME"
)

type PasswordCommand struct {
	*base.Command

	flagLoginName string
	flagPassword  string
}

func (c *PasswordCommand) Synopsis() string {
	return wordwrap.WrapString("Invoke the password auth method to authenticate with Boundary", base.TermWidth)
}

func (c *PasswordCommand) Help() string {
	return base.WrapForHelpText([]string{
		"Usage: boundary authenticate password [options] [args]",
		"",
		"  Invoke the password auth method to authenticate the Boundary CLI. Example:",
		"",
		`    $ boundary authenticate password -auth-method-id ampw_1234567890 -login-name foo`,
		"",
		"",
	}) + c.Flags().Help()
}

func (c *PasswordCommand) Flags() *base.FlagSets {
	set := c.FlagSet(base.FlagSetHTTP | base.FlagSetClient | base.FlagSetOutputFormat)
	f := set.NewFlagSet("Command Options")

	f.StringVar(&base.StringVar{
		Name:   "login-name",
		Target: &c.flagLoginName,
		EnvVar: envLoginName,
		Usage:  "The login name corresponding to an account within the given auth method",
	})

	f.StringVar(&base.StringVar{
		Name:   "password",
		Target: &c.flagPassword,
		EnvVar: envPassword,
		Usage:  "The password associated with the login name",
	})

	f.StringVar(&base.StringVar{
		Name:   "auth-method-id",
		EnvVar: "BOUNDARY_AUTH_METHOD_ID",
		Target: &c.FlagAuthMethodId,
		Usage:  "The auth-method resource to use for the operation",
	})

	return set
}

func (c *PasswordCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything
}

func (c *PasswordCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

type dummyGenericResponse struct {
	item     interface{}
	response *api.Response
}

var _ api.GenericResult = (*dummyGenericResponse)(nil)

func (d *dummyGenericResponse) GetItem() interface{} {
	return d.item
}

func (d *dummyGenericResponse) GetResponse() *api.Response {
	return d.response
}

func (c *PasswordCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.PrintCliError(err)
		return base.CommandUserError
	}

	switch {
	case c.flagLoginName == "":
		c.PrintCliError(errors.New("Login name must be provided via -login-name"))
		return base.CommandUserError
	case c.FlagAuthMethodId == "":
		c.PrintCliError(errors.New("Auth method ID must be provided via -auth-method-id"))
		return base.CommandUserError
	}

	if c.flagPassword == "" {
		fmt.Print("Password is not set as flag or in env, please enter it now (will be hidden): ")
		value, err := password.Read(os.Stdin)
		fmt.Print("\n")
		if err != nil {
			c.UI.Error(fmt.Sprintf("An error occurred attempting to read the password. The raw error message is shown below but usually this is because you attempted to pipe a value into the command or you are executing outside of a terminal (TTY). The raw error was:\n\n%s", err.Error()))
			return base.CommandUserError
		}
		c.flagPassword = strings.TrimSpace(value)
	}

	client, err := c.Client(base.WithNoTokenScope(), base.WithNoTokenValue())
	if c.WrapperCleanupFunc != nil {
		defer func() {
			if err := c.WrapperCleanupFunc(); err != nil {
				c.PrintCliError(fmt.Errorf("Error cleaning kms wrapper: %w", err))
			}
		}()
	}
	if err != nil {
		c.PrintCliError(fmt.Errorf("Error creating API client: %w", err))
		return base.CommandCliError
	}

	aClient := authmethods.NewClient(client)
	result, err := aClient.Authenticate(c.Context, c.FlagAuthMethodId, "login",
		map[string]interface{}{
			"login_name": c.flagLoginName,
			"password":   c.flagPassword,
		})
	if err != nil {
		if apiErr := api.AsServerError(err); apiErr != nil {
			c.PrintApiError(apiErr, "Error from controller when performing authentication")
			return base.CommandApiError
		}
		c.PrintCliError(fmt.Errorf("Error trying to perform authentication: %w", err))
		return base.CommandCliError
	}

	return saveAndOrPrintToken(c.Command, result)
}
