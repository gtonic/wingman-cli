package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

type Command = cli.Command

type Flag = cli.Flag
type IntFlag = cli.IntFlag
type IntSliceFlag = cli.IntSliceFlag
type StringFlag = cli.StringFlag
type StringSliceFlag = cli.StringSliceFlag
type BoolFlag = cli.BoolFlag

func ShowAppHelp(cmd *Command) error {
	return cli.ShowAppHelp(cmd)
}

func ShowCommandHelp(cmd *Command) error {
	return cli.ShowSubcommandHelp(cmd)
}

func Prompt(label, placeholder string) (string, error) {
	var s string

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stderr, label+" ")

		s, _ = r.ReadString('\n')

		if s != "" {
			break
		}
	}

	return strings.TrimSpace(s), nil
}

func MustPrompt(label, placeholder string) string {
	value, err := Prompt(label, placeholder)

	if err != nil {
		panic(err)
	}

	return value
}
func Confirm(label string, placeholder bool) (bool, error) {
	choices := "Y/n"

	if !placeholder {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)

	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)

		if s == "" {
			return placeholder, nil
		}

		s = strings.ToLower(s)

		if s == "y" || s == "yes" {
			return true, nil
		}

		if s == "n" || s == "no" {
			return false, nil
		}
	}
}

func MustConfirm(label string, placeholder bool) bool {
	value, err := Confirm(label, placeholder)

	if err != nil {
		panic(err)
	}

	return value
}
