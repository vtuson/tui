/* This is a Basic Implementation of a text ui that can be used to run CLI commands or you can define your own HandlerCommand to run Go functions
  Basic Structure is:
	Menu
 		[]Command
 			[]Args

 By default a CLI command handler is provided that is able to pass parameters as flag, options or Enviroment variables

 You can find an example app in the /example folder
*/

package tui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"os"
	"os/exec"
)

// Defines a basic Command object
// HandlerCommand function can be defined custom, but if not defined the code defaults to an OSHandler that will
// Execute the command path under Cli and pass the args as flags,envar or values
// Optional is currently WIP

type Command struct {
	Title       string
	Description string
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
}

//Argument can be a flag (IsFlag) or a Envar (if defined). If IsFalg is false the Name is passed without a - appended
//IsBoolean means that the argument is passed with no additional value
// flag bool: -foo
// flag: -foo bar
// noflag bool: foo
type Argument struct {
	Envar       string
	Name        string
	IsFlag      bool
	Title       string
	Description string
	Value       string
	IsBoolean   bool
	Valuebool   bool
}

//Top level menu description
//Bottombar provides information to the user on how to exit the menu
//All *Text have default english values but can be overwritten

type Menu struct {
	Title         string
	Description   string
	Commands      []Command
	Cursor        int
	BottomBar     bool
	BottomBarText string   //text for default top menu
	BackText      string   //text for back text on command
	BoolText      string   //text when an arg is a bool
	ValueText     string   //text when an arg is a value string
	Wait          chan int //channel to wait completion
	p             *Printing
	breadCrum     string
	argIndex      int
	enableScape   bool
	runeBuffer    []rune
}

//gets a breadcrum for a command
func (c *Command) BreadCrum() string {
	return c.breadCrum + " > " + c.Title
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
type HandlerCommand func(c *Command, screen chan string)

func (m *Menu) SelectToggle() {
	if m.Cursor < len(m.Commands) {
		m.Commands[m.Cursor].Selected = !m.Commands[m.Cursor].Selected
	}
}
func (m *Menu) IsToggle() bool {
	if m.Cursor < len(m.Commands) {
		return m.Commands[m.Cursor].Optional
	}
	return false
}

func (m *Menu) printPageHearder(title string, desc string) {
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
func OSCmdHandler(c *Command, ch chan string) {
	formattedArgs := []string{"test"}
	c.Error = nil

	defer close(ch)
	for _, a := range c.Args {
		if a.Envar != "" {
			if a.IsBoolean {
				if a.Valuebool {
					os.Setenv(a.Envar, "true")
				} else {
					os.Unsetenv(a.Envar)
				}
			} else {
				os.Setenv(a.Envar, a.Value)
			}
			continue
		}
		if a.IsBoolean {
			if a.Valuebool {
				if a.IsFlag {
					formattedArgs = append(formattedArgs, "-"+a.Name)
				} else {
					formattedArgs = append(formattedArgs, a.Name)
				}
			} else {
				continue
			}
		} else {
			formattedArgs = append(formattedArgs, "-"+a.Name, a.Value)
		}
	}

	cmd := exec.Command(c.Cli, formattedArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.Error = err
		return
	}
	c.Error = cmd.Start()
	if c.Error != nil {
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

	c.Error = cmd.Wait()
}

//Run commands and waits to complete, then calls menu ShowResult
func (m *Menu) RunCommand(c *Command) {
	if c.Execute == nil {
		c.Execute = OSCmdHandler
	}
	ch := make(chan string)
	go c.Execute(c, ch)

	pb := NewProgressBar(m.p)
	go pb.Start()

	for ok := true; ok; {
		_, ok = <-ch
	}
	pb.Stop()
	m.ShowResult(c)
}

//Displays result of running a command, using test Fail and Success, plus adds error message for Fail
func (m *Menu) ShowResult(c *Command) {
	m.enableScape = true

	if c.Error != nil {
		m.p.Clear()
		m.printPageHearder(c.BreadCrum(), c.Fail+" Error ocurred:"+c.Error.Error())
	} else {
		m.p.Clear()
		m.printPageHearder(c.BreadCrum(), "Success! "+c.Success)
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
	m.ShowCommand()
}

//Displays a command in the screen as incated by Cursor in Menu
func (m *Menu) ShowCommand() {
	if m.Cursor >= len(m.Commands) {
		fmt.Println("Cursor exceed array")
		return
	}
	c := m.Commands[m.Cursor]
	m.p.Clear()

	runNow := (c.Args == nil || len(c.Args) == 0 || m.argIndex >= len(c.Args))

	c.breadCrum = m.BreadCrum()
	if !runNow {
		m.printPageHearder(c.BreadCrum()+" > "+c.Args[m.argIndex].Name, c.Description)
		m.p.Putln(c.Args[m.argIndex].Title+":", false)
		m.p.Putln(c.Args[m.argIndex].Description, false)
		if c.Args[m.argIndex].IsBoolean {
			m.p.BottomBar(m.BoolText)
		} else {
			m.p.BottomBar(m.ValueText)
			m.p.PutEcho(string([]rune{tcell.RuneBlock}), m.p.style.Input)
		}

	} else {
		m.enableScape = false
		m.printPageHearder(c.BreadCrum(), c.Description)
	}

	m.p.Show()

	if runNow {
		m.RunCommand(&c)
	}

}

//Show Menu
func (m *Menu) Show() {
	m.p.Clear()
	m.printPageHearder(m.BreadCrum(), m.Description)
	for i, c := range m.Commands {
		title := c.Title
		if c.Optional {
			check := "[ ]"
			if c.Selected {
				check = "[x]"
			}
			title = check + " " + title
		}
		status := ""
		if c.Status != "" {
			status = " [" + c.Status + "]"
		}
		if c.Disable && i != m.Cursor {
			m.p.PutlnDisable(title + status)
			continue
		}
		m.p.Putln(title+status, i == m.Cursor)
	}
	if m.BottomBar {
		m.p.BottomBar(m.BottomBarText)
	}
	m.p.Show()
}
func (m *Menu) CurrentCommand() *Command {
	if m.Cursor < len(m.Commands) {
		return &m.Commands[m.Cursor]
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
	if m.Cursor < len(m.Commands)-1 {
		m.Cursor++
	} else {
		m.Cursor = 0
	}
	m.p.Clear()
	m.Show()
}

//Moves back to the prev command
func (m *Menu) Prev() {
	if m.Cursor > 0 {
		m.Cursor--
	} else {
		m.Cursor = len(m.Commands) - 1
	}
	m.p.Clear()
	m.Show()
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
		BoolText:      "Press Y for yes or N for No, ESC to Cancel",
		ValueText:     "Type your answer and press ENTER to continue, or ESC to Cancel",
		Wait:          channel,
		p:             p,
		runeBuffer:    []rune{},
	}
}

//Handles key events for commnands
func (menu *Menu) EventCommandManager() {
	for {
		ev := menu.p.Screen().PollEvent()
		cmd := menu.CurrentCommand()
		var arg *Argument
		if cmd != nil && menu.argIndex < len(cmd.Args) {
			arg = &cmd.Args[menu.argIndex]
		}
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				if menu.enableScape {
					menu.Show()
					go menu.EventManager()
					return
				}
			case tcell.KeyEnter:
				if arg != nil && !arg.IsBoolean {
					arg.Value = string(menu.runeBuffer)
					menu.NextArgument()
				}
			case tcell.KeyCtrlL:
				menu.p.Sync()
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if arg != nil && !arg.IsBoolean {
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
				close(menu.Wait)
				return
			case tcell.KeyEnter:
				cmd := menu.CurrentCommand()
				if cmd != nil && cmd.Disable {
					continue
				}
				if menu.IsToggle() {
					menu.SelectToggle()
					continue
				}
				menu.argIndex = 0
				menu.enableScape = true
				menu.ShowCommand()
				go menu.EventCommandManager()
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
