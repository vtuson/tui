package tui

import (
	"testing"
)

func TestNoArgsPass(t *testing.T) {
	tmpc := Command{
		Title:       "No Args",
		Cli:         "./sampleapp/testcommands/waitqok.sh",
		Description: "test of running a Command with out arguments",
		Success:     "Yey it works",
		Execute:     OSCmdHandler,
	}
	ch := make(chan string)
	go tmpc.Execute(&tmpc, ch)
	for ok := true; ok; {
		_, ok = <-ch
	}
	if tmpc.Error != nil {
		t.Log(tmpc.Error)
		t.Fail()
	}
}

func TestNoArgsFail(t *testing.T) {
	tmpc := Command{
		Title:       "foo",
		Cli:         "./sampleapp/testcommands/args.sh",
		Description: "foo",
		Success:     "foo",
		Execute:     OSCmdHandler,
	}
	ch := make(chan string)
	go tmpc.Execute(&tmpc, ch)
	for ok := true; ok; {
		_, ok = <-ch
	}
	if tmpc.Error == nil {
		t.Log(tmpc.Error)
		t.Fail()
	}
}

func TestArgsPass(t *testing.T) {
	tmpc := Command{
		Title:       "Args CLI",
		Cli:         "./sampleapp/testcommands/args.sh",
		Description: "test of running a tui.Command with arguments",
		Args: []Argument{
			Argument{
				Description: "Would you like to set this is a sample flag bool?",
				Title:       "Sample Flag Bool",
				IsBoolean:   true,
				Name:        "first",
				IsFlag:      true,
				Valuebool:   true,
			},
			Argument{
				Description: "This is a sample value flag",
				Title:       "Sample Flag with value",
				Name:        "second",
				IsFlag:      true,
				Value:       "foo",
			},
			Argument{
				Description: "Would you like to set this is a sample bool?",
				Title:       "Sample Bool",
				IsBoolean:   true,
				Name:        "third",
				Valuebool:   true,
			},
		},
		Success: "Yey it works",
		Fail:    "oh, it didnt work.",
		Execute: OSCmdHandler,
	}
	ch := make(chan string)
	go tmpc.Execute(&tmpc, ch)
	for ok := true; ok; {
		_, ok = <-ch
	}
	if tmpc.Error != nil {
		t.Log(tmpc.Error)
		t.Fail()
	}
}

func TestArgsEnvPass(t *testing.T) {
	tmpc := Command{
		Title:       "Args Envar",
		Cli:         "./sampleapp/testcommands/env.sh",
		Description: "test of running a command with arguments via envar",
		Args: []Argument{
			Argument{
				Description: "Would you like to set this is a sample Envar bool?",
				Title:       "Sample Envar Bool",
				IsBoolean:   true,
				Name:        "first",
				IsFlag:      true,
				Envar:       "FIRSTTEST",
				Valuebool:   true,
			},
		},
		Success: "Yey it works",
		Fail:    "oh, it didnt work.",
		Execute: OSCmdHandler,
	}
	ch := make(chan string)
	go tmpc.Execute(&tmpc, ch)
	for ok := true; ok; {
		_, ok = <-ch
	}
	if tmpc.Error != nil {
		t.Log(tmpc.Error)
		t.Fail()
	}
}
