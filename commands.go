package tui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"os"
	"os/exec"
)

/*
This is a Basic Implementation of a text ui that can be used to run CLI commands or you can define your own HandlerCommand to run Go functions
Basic Structure is:
	Menu
		[]Command
			[]Args

By default a CLI command handler is provided that is able to pass parameters as flag, options or Enviroment variables
*/
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

type Menu struct {
	Title         string
	Description   string
	Commands      []Command
	Cursor        int
	BottomBar     bool
	BottomBarText string
	BackText      string
	BoolText      string
	ValueText     string
	Wait          chan int
	p             *Printing
	breadCrum     string
	argIndex      int
	enableScape   bool
	runeBuffer    []rune
}

func (c *Command) BreadCrum() string {
	return c.breadCrum + " > " + c.Title
}

func (m *Menu) BreadCrum() string {
	path := ""
	if m.breadCrum != "" {
		path = m.breadCrum + " > "
	}
	return path + m.Title
}

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
func (m *Menu) NextCommand() {
	m.argIndex++
	m.runeBuffer = []rune{}
	m.ShowCommand()
}

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

func (m *Menu) Quit() {
	m.p.s.Fini()
}

func (m *Menu) Next() {
	if m.Cursor < len(m.Commands)-1 {
		m.Cursor++
	} else {
		m.Cursor = 0
	}
	m.p.Clear()
	m.Show()
}
func (m *Menu) Prev() {
	if m.Cursor > 0 {
		m.Cursor--
	} else {
		m.Cursor = len(m.Commands) - 1
	}
	m.p.Clear()
	m.Show()
}

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
					menu.NextCommand()
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
						menu.NextCommand()
						break
					}
					if ev.Rune() == 'N' || ev.Rune() == 'n' {
						arg.Valuebool = false
						menu.NextCommand()
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
