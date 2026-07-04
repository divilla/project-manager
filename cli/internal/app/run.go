package app

import (
	"errors"
	"flag"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Version is the user-visible mch executable version.
const Version = "0.1"

// Run executes the mch command with the supplied process arguments and output writer.
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

	lipgloss.SetColorProfile(termenv.ANSI256)
	_, err := tea.NewProgram(NewModel(), tea.WithOutput(out)).Run()
	return err
}
