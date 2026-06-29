package app

import (
	"errors"
	"flag"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

const Version = "0.1"

func Run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("mch", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	showVersion := fs.Bool("version", false, "print version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("mch does not accept subcommands")
	}
	if *showVersion {
		_, err := fmt.Fprintf(out, "mch %s\n", Version)
		return err
	}

	_, err := tea.NewProgram(NewModel(), tea.WithOutput(out)).Run()
	return err
}
