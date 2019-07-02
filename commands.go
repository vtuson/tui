// This is a Basic Implementation of a text ui that can be used to run CLI commands
// or you can define your own HandlerCommand to run Go functions
// Basic Structure is:
// Menu
//    []Option
//      - SubOptions or Command
//
// By default a CLI command handler is provided that is able to pass parameters as flag,
// options or Enviroment variables
//
// You can find an example app in the /example folder
//
package tui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"os"
	"os/exec"
	"strings"
)

var arrayMenu []Menu

// TODO
// Defines a basic Command object
// HandlerCommand function can be defined custom, but if not defined the code defaults to an OSHandler that will
// Execute the command path under Cli and pass the args as flags,envar or values
// Optional is currently WIP

type Option struct {
	Title       string
	Description string
  SubOptions  []Option
	Cli         string
	Execute     HandlerCommand
	Args        []Argument
	Optional    bool
	Selected    bool
	Error       error
	Success     string
	Fail        string
	breadCrum   string
	Disable     bool
	Status      string
	bufferOut   []string
	PrintOut    bool
	External    bool
}

//TODO
//Argument can be a flag (IsFlag) or a Envar (if defined). If IsFalg is false the Name is passed without a - appended
//IsBoolean means that the argument is passed with no additional value
// flag bool: -foo
// flag: -foo bar
// noflag bool: foo
type Argument struct {
	Envar       string
	Name        string
	Title       string
	Description string
	Value       []string
	IsBoolean   bool
	Valuebool   bool
  Multiple    bool
}

//Top level menu description
//Bottombar provides information to the user on how to exit the menu
//All *Text have default english values but can be overwritten

type Menu struct {
	Title         string
	Description   string
	Options       []Option
	Cursor        int
	BottomBar     bool
	BottomBarText string   //text for default top menu
	BackText      string   //text for back text on command
	BoolText      string   //text when an arg is a bool
	ValueText     string   //text when an arg is a value string
	ValueTextMultiple string   //text when an arg is a value string and accept multiple values
	Wait          chan int //channel to wait completion
	p             *Printing
	breadCrum     string
	argIndex      int
	enableScape   bool
	runeBuffer    []rune
}

//gets a breadcrum for a command
func (o *Option) BreadCrum(m Menu) string {
	return m.BreadCrum() + " > " + o.Title
}

//gets a breadcrum for a menus
func (m *Menu) BreadCrum() string {
	path := ""
	if m.breadCrum != "" {
		path = m.breadCrum + " > "
	}
	return path + m.Title
}

// type def for a handler function, pass a command and a channel to return screen stdout, channel will be close when
// execution is completed
type HandlerCommand func(o *Option, screen chan string)

func (m *Menu) SelectToggle() {
	if m.Cursor < len(m.Options) {
		m.Options[m.Cursor].Selected = !m.Options[m.Cursor].Selected
	}
}
func (m *Menu) IsToggle() bool {
	if m.Cursor < len(m.Options) {
		return m.Options[m.Cursor].Optional
	}
	return false
}

func (m *Menu) printPageHeader(title string, desc string) {
	if title != "" {
		m.p.Putln(title, true)
		m.p.Return()
	}

	if desc != "" {
		m.p.Putln(desc, false)
		m.p.Return()
	}
}

//OS execution default handler
//Error in command is updated in completion
func OSCmdHandler(o *Option, ch chan string) {
	formattedArgs := []string{}
	o.Error = nil
	cliArray := strings.Split(o.Cli, " ")
	formattedArgs = cliArray[1:]

	defer close(ch)
	for _, a := range o.Args {
		if a.Envar != "" {
			if a.IsBoolean {
				if a.Valuebool {
					os.Setenv(a.Envar, "true")
				} else {
					os.Unsetenv(a.Envar)
				}
			} else {
				os.Setenv(a.Envar, a.Value[0])
			}
			continue
		}
		if a.IsBoolean {
			if a.Valuebool {
        formattedArgs = append(formattedArgs, a.Name)
			} else {
				continue
			}
		} else {
      for _, argumentValue := range a.Value {
        formattedArgs = append(formattedArgs, a.Name, argumentValue)
      }
		}
	}

	cmd := exec.Command(cliArray[0], formattedArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		o.Error = err
		return
	}
	o.Error = cmd.Start()
	if o.Error != nil {
		return
	}
	end := false
	for !end {
		p := make([]byte, 1)
		if _, err := stdout.Read(p); err == nil {
			ch <- string(p)
		} else {
			end = true
		}
	}

	o.Error = cmd.Wait()
}

//Run commands and waits to complete, then calls menu ShowResult
func (m *Menu) RunCommand(o *Option) {
	if o.External {
    close(menu.Wait)
    return
	}
	if o.Execute == nil {
		o.Execute = OSCmdHandler
	}
	if o.PrintOut {
		o.bufferOut = []string{}
	}
	ch := make(chan string)
	go o.Execute(o, ch)

	pb := NewProgressBar(m.p)
	go pb.Start()

	for ok := true; ok; {
		var b string
		b, ok = <-ch
		if ok {
			o.bufferOut = append(o.bufferOut, b)
		}
	}
	pb.Stop()
	m.ShowResult(o)
}

//Displays result of running a command, using test Fail and Success, plus adds error message for Fail
func (m *Menu) ShowResult(o *Option) {

	if o.Error != nil {
		m.p.Clear()
		m.printPageHeader(o.BreadCrum(*m), o.Fail+" Error ocurred:"+o.Error.Error())
	} else {
		m.p.Clear()
		m.printPageHeader(o.BreadCrum(*m), o.Success)
	}

	if o.PrintOut {
		tmpOut := strings.Join(o.bufferOut, "")
		tmpOut = strings.Replace(tmpOut, "\r", "\n", -1)
		output := strings.Split(tmpOut, "\n")
		for _, l := range output {
			m.p.Putln(l, false)
		}
	}

	if m.BottomBar {
		m.p.BottomBar(m.BackText)
	}
	m.p.Show()

}

//Moves menu to the Next Argument in a Command
func (m *Menu) NextArgument() {
	m.argIndex++
	m.runeBuffer = []rune{}
	m.ShowOption()
}

//Displays a command in the screen as incated by Cursor in Menu
func (m *Menu) ShowOption() {
	if m.Cursor >= len(m.Options) {
		fmt.Println("Cursor exceed array")
		return
	}
	o := m.Options[m.Cursor]
	m.p.Clear()

  hasSubmenu := o.SubOptions != nil && len(o.SubOptions) > 0
	runCommandNow := (o.Args == nil || len(o.Args) == 0 || m.argIndex >= len(o.Args))

  if hasSubmenu {
    arrayMenu = append(arrayMenu, *m)
    m.Title = fmt.Sprintf("%s > %s", m.Title, o.Title)
    m.Description = o.Description
    m.Options = o.SubOptions
		m.enableScape = true
		m.Cursor = 0
    m.PrintMenu()
    go m.EventManager()
  } else if !runCommandNow {
		m.printPageHeader(o.BreadCrum(*m)+" > "+o.Args[m.argIndex].Name, o.Description)
		m.p.Putln(o.Args[m.argIndex].Title+":", false)
		m.p.Putln(o.Args[m.argIndex].Description, false)
		if o.Args[m.argIndex].IsBoolean {
			m.p.BottomBar(m.BoolText)
		} else {
      if o.Args[m.argIndex].Multiple {
        m.p.BottomBar(m.ValueTextMultiple)
      } else {
        m.p.BottomBar(m.ValueText)
      }
			m.p.PutEcho(string([]rune{tcell.RuneBlock}), m.p.style.Input)
		}
  } else {
		m.printPageHeader(o.BreadCrum(*m), o.Description)
  }

  if !hasSubmenu {
    m.p.Show()
    if runCommandNow {m.RunCommand(&o)}
  }
}

//Show Menu
func (m *Menu) PrintMenu() {
	m.p.Clear()
	m.printPageHeader(m.BreadCrum(), m.Description)
	for i, o := range m.Options {
		title := o.Title
		if o.Optional {
			check := "[ ]"
			if o.Selected {
				check = "[x]"
			}
			title = check + " " + title
		}
		status := ""
		if o.Status != "" {
			status = " [" + o.Status + "]"
		}
		if o.Disable && i != m.Cursor {
			m.p.PutlnDisable(title + status)
			continue
		}
		m.p.Putln(title+status, i == m.Cursor)
	}
	if m.BottomBar {
    if m.enableScape {
      m.p.BottomBar(m.BackText)
    } else {
      m.p.BottomBar(m.BottomBarText)
    }
	}
	m.p.Show()
}
func (m *Menu) CurrentOption() *Option {
	if m.Cursor < len(m.Options) {
		return &m.Options[m.Cursor]
	} else {
		return nil
	}

}

//Leaves a menu
func (m *Menu) Quit() {
	m.p.s.Fini()
}

//Moves to the next command in a list
func (m *Menu) Next() {
	if m.Cursor < len(m.Options)-1 {
		m.Cursor++
	} else {
		m.Cursor = 0
	}
	m.p.Clear()
	m.PrintMenu()
}

//Moves back to the prev command
func (m *Menu) Prev() {
	if m.Cursor > 0 {
		m.Cursor--
	} else {
		m.Cursor = len(m.Options) - 1
	}
	m.p.Clear()
	m.PrintMenu()
}

//Returns an initialised default Menu
func NewMenu(style *Style) *Menu {
	channel := make(chan int)
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	p := NewPrinting(s, style)
	return &Menu{
		BottomBar:     true,
		BottomBarText: "Press ESC to exit",
		BackText:      "Press ESC to go back",
		BoolText:      "Press Y for yes or N for No, ESC to cancel",
		ValueText:     "Type your answer and press ENTER to continue, or ESC to cancel",
		ValueTextMultiple: "Type your answer and press ENTER to continue, or ESC to cancel. Set an empty string to stop adding values",
		Wait:          channel,
		p:             p,
		runeBuffer:    []rune{},
	}
}

//Handles key events for commnands
func (menu *Menu) EventCommandManager() {
	for {
		ev := menu.p.Screen().PollEvent()
		o := menu.CurrentOption()
		var arg *Argument
		if o != nil && menu.argIndex < len(o.Args) {
			arg = &o.Args[menu.argIndex]
		}
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
        menu.PrintMenu()
        go menu.EventManager()
        return
			case tcell.KeyEnter:
				if arg != nil && !arg.IsBoolean {
          userInput := string(menu.runeBuffer)
          if userInput == "" {
            if arg.Multiple && len(arg.Value) > 0 {
              menu.NextArgument()
            } else {
              menu.runeBuffer = []rune{}
              menu.ShowOption()
            }
          } else {
            arg.Value = append(arg.Value, userInput)
            if arg.Multiple {
              menu.runeBuffer = []rune{}
              menu.ShowOption()
            } else {
              menu.NextArgument()
            }
          }
				}
			case tcell.KeyCtrlL:
				menu.p.Sync()
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if arg != nil && !arg.IsBoolean && len(menu.runeBuffer) > 0 {
					menu.runeBuffer = menu.runeBuffer[:len(menu.runeBuffer)-1]
					menu.p.PutEcho(string(append(menu.runeBuffer, tcell.RuneBlock)), menu.p.style.Input)
					menu.p.Show()
				}
			case tcell.KeyRune:
				if arg != nil && arg.IsBoolean {
					if ev.Rune() == 'y' || ev.Rune() == 'Y' {
						arg.Valuebool = true
						menu.NextArgument()
						break
					}
					if ev.Rune() == 'N' || ev.Rune() == 'n' {
						arg.Valuebool = false
						menu.NextArgument()
						break
					}
				} else {
					menu.runeBuffer = append(menu.runeBuffer, ev.Rune())
					menu.p.PutEcho(string(append(menu.runeBuffer, tcell.RuneBlock)), menu.p.style.Input)
					menu.p.Show()
				}
			}

		case *tcell.EventResize:
			menu.p.Sync()
		}
	}
}

//Handles Key events for Menu
func (menu *Menu) EventManager() {
	for {
		ev := menu.p.Screen().PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
        if menu.enableScape {
          *menu, arrayMenu = arrayMenu[len(arrayMenu)-1], arrayMenu[:len(arrayMenu)-1]
          menu.argIndex = 0
          menu.PrintMenu()
          go menu.EventManager()
        } else {
          close(menu.Wait)
        }
				return
			case tcell.KeyEnter:
				o := menu.CurrentOption()
				if o != nil && o.Disable {
					continue
				}
				if menu.IsToggle() {
					menu.SelectToggle()
					continue
				}
				menu.argIndex = 0
				menu.ShowOption()
        if o.SubOptions == nil || len(o.SubOptions) == 0 {
          go menu.EventCommandManager()
        } 
				return

			case tcell.KeyCtrlL:
				menu.p.Sync()
			case tcell.KeyUp:
				menu.Prev()
				menu.p.Show()
			case tcell.KeyDown:
				menu.Next()
				menu.p.Show()
			}
		case *tcell.EventResize:
			menu.p.Sync()
		}
	}
}
