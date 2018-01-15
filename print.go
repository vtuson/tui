package tui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mattn/go-runewidth"
	"os"
)

var styleText = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.Color17).Bold(true)
var styleTextHighlight = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.Color17).Bold(true)
var styleMenu = tcell.StyleDefault.Background(tcell.ColorSilver).Foreground(tcell.Color17).Bold(true)
var styleInput = tcell.StyleDefault.Foreground(tcell.ColorLime).Background(tcell.Color17).Bold(false)
var styleDisable = tcell.StyleDefault.Foreground(tcell.ColorSilver).Background(tcell.Color17).Bold(false)

//Defines style to be used on a menu
// Default styles are provide
// Indent is the number of spaces from the right side
type Style struct {
	Indent     int
	Hightlight tcell.Style
	Default    tcell.Style
	Menu       tcell.Style
	H1         tcell.Style
	Input      tcell.Style
	Disable    tcell.Style
}

//returns default style
func DefaultStyle() *Style {
	return &Style{
		Hightlight: styleTextHighlight,
		Default:    styleText,
		Menu:       styleMenu,
		H1:         styleTextHighlight,
		Input:      styleInput,
		Disable:    styleDisable,
		Indent:     2,
	}
}

//returns printing object
func NewPrinting(s tcell.Screen, style *Style) *Printing {
	encoding.Register()

	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.Color17))
	s.Clear()

	return &Printing{
		s:     s,
		style: style,
	}
}

type Printing struct {
	s       tcell.Screen
	Cursor  int
	xcursor int
	style   *Style
}

//Clears screen and goes back to the top
func (p *Printing) Clear() {
	p.s.Clear()
	p.Top()
}

//gets the screen object
func (p *Printing) Screen() tcell.Screen {
	return p.s
}

//shows the screen
func (p *Printing) Show() {
	p.s.Show()
}

//syncs the screen
func (p *Printing) Sync() {
	p.s.Sync()
}

//moves cursor to the top
func (p *Printing) Top() {
	p.Cursor = 0
}

//moves cursor to the last line
func (p *Printing) Bottom() {
	_, p.Cursor = p.s.Size()
	p.Cursor--
}

//add blank line
func (p *Printing) Return() {
	p.Cursor++
}

//adds a ln with a return at the end. hightlight uses the highlight style for the full line text
func (p *Printing) Putln(str string, highlight bool) {
	if highlight {
		p.Cursor = p.puts(p.style.Hightlight, p.style.Indent, p.Cursor, str)
	} else {
		p.Cursor = p.puts(p.style.Default, p.style.Indent, p.Cursor, str)
	}
	p.Cursor++
}

//add a line disabled style
func (p *Printing) PutlnDisable(str string) {
	p.Cursor = p.puts(p.style.Disable, p.style.Indent, p.Cursor, str)
	p.Cursor++
}

//same as Putln but continues on x
func (p *Printing) Put(str string, highlight bool) {
	if highlight {
		p.Cursor = p.puts(p.style.Hightlight, p.style.Indent+p.xcursor, p.Cursor, str)
	} else {
		p.Cursor = p.puts(p.style.Default, p.style.Indent+p.xcursor, p.Cursor, str)
	}
}

//prints string but does not move cursor
func (p *Printing) PutEcho(str string, style tcell.Style) {

	p.puts(style, p.style.Indent+p.xcursor, p.Cursor, str)
}

//putsln with H1 style
func (p *Printing) PutH1(str string, highlight bool) {
	p.Cursor = p.puts(p.style.H1, p.style.Indent, p.Cursor, str)
	p.Cursor++
}

//prints the bottome bar
func (p *Printing) BottomBar(str string) {
	_, y := p.s.Size()
	p.puts(p.style.Menu, 0, y-1, "  "+str)
}

//puts a string line without filling the end spaces
func (p *Printing) putc(style tcell.Style, x, y int, str string) int {
	i := 0
	var deferred []rune
	dwidth := 0
	xScreen, _ := p.s.Size()

	for _, r := range str {
		if x+i >= xScreen-p.style.Indent {
			i = 0
			y++
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				p.s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				p.s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		p.s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
	return y
}

//puts a line filling the trailing spaces to the end with the same style
func (p *Printing) puts(style tcell.Style, x, y int, str string) int {
	i := 0
	var deferred []rune
	dwidth := 0
	xScreen, _ := p.s.Size()

	for _, r := range str {
		if x+i >= xScreen-p.style.Indent {
			i = 0
			y++
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				p.s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				p.s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		p.s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
	for i < xScreen {
		p.s.SetContent(x+i, y, ' ', nil, style)
		i++
	}
	return y

}
