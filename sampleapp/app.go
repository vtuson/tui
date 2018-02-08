package main

import (
	"github.com/vtuson/tui"
)

func NewTestMenu() *tui.Menu {
	m := tui.NewMenu(tui.DefaultStyle())
	m.Title = "Test"
	m.Description = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas id congue felis,vitae auctor metus. Morbi placerat lectus a velit feugiat, ac tincidunt ex ultricies. Nullam fermentum vestibulum tellus, gravida lacinia dui fringilla eget. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus.`

	tmpcs := []tui.Command{
		tui.Command{
			Title:       "No Args",
			Cli:         "./testcommands/waitok.sh",
			Description: "test of running a tui.Command with out arguments",
			Success:     "Yey it works",
			PrintOut:    true,
		},
		tui.Command{
			Title:       "Commmand Failing",
			Cli:         "./testcommands/args.sh",
			Description: "test of running a that returns exit 1",
			Success:     "Yey it works",
			Fail:        "oh, it didnt work.",
		},
		tui.Command{
			Title:       "Args CLI",
			Cli:         "./testcommands/args.sh",
			Description: "test of running a tui.Command with arguments",
			Args: []tui.Argument{
				tui.Argument{
					Description: "Would you like to set this is a sample flag bool?",
					Title:       "Sample Flag Bool",
					IsBoolean:   true,
					Name:        "first",
					IsFlag:      true,
				},
				tui.Argument{
					Description: "This is a sample value flag",
					Title:       "Sample Flag with value",
					Name:        "second",
					IsFlag:      true,
				},
				tui.Argument{
					Description: "Would you like to set this is a sample bool?",
					Title:       "Sample Bool",
					IsBoolean:   true,
					Name:        "third",
				},
			},
			Success: "Yey it works",
			Fail:    "oh, it didnt work.",
		},
		tui.Command{
			Title:       "Args Envar",
			Cli:         "./testcommands/env.sh",
			Description: "test of running a command with arguments via envar",
			Args: []tui.Argument{
				tui.Argument{
					Description: "Would you like to set this is a sample Envar bool?",
					Title:       "Sample Envar Bool",
					IsBoolean:   true,
					Name:        "first",
					IsFlag:      true,
					Envar:       "FIRSTTEST",
				},
			},
			Success: "Yey it works",
			Fail:    "oh, it didnt work.",
		},
		tui.Command{
			Title:   "Done cmd",
			Disable: true,
			Status:  "Done",
		},
	}
	m.Commands = tmpcs
	return m
}

func main() {
	menu := NewTestMenu()

	menu.Show()
	go menu.EventManager()
	<-menu.Wait
	menu.Quit()
}
