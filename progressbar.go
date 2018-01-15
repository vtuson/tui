package tui

import (
	"fmt"
	"time"
)

type ProgressBar struct {
	Interval int
	p        *Printing
	ticker   *time.Ticker
	Text     string
}

func NewProgressBar(p *Printing) *ProgressBar {
	return &ProgressBar{p: p, Interval: 500, Text: "Please wait"}
}

func (p *ProgressBar) Start() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(p.Interval))
	p.ticker = ticker
	x, y := p.p.Screen().Size()

	p.p.putc(p.p.style.Default, x/2-len(p.Text)/2+1, y/2-1, p.Text)
	i := 0
	bar := []string{" | ", " / ", " - ", " \\ "}
	for ok := true; ok; {
		if i >= len(bar) {
			i = 0
		}
		p.p.putc(styleBitnamiTextHighlight, x/2, y/2, bar[i])
		p.p.Show()
		i++
		_, ok = <-ticker.C
	}
	fmt.Println("exit")

}

func (p *ProgressBar) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
