package main

import (
	"image/color"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"

	"golang.org/x/image/font/inconsolata"

	"github.com/tapir/convar"
)

type game struct {
	console     *convar.Console
	command     string
	prevCommand string
	counter     int
	complete    []string
}

func (g *game) Update() error {
	// Add a string from InputChars, that returns string input by users
	if len([]rune(g.command)) < 90 {
		g.command += string(ebiten.InputChars())
	}

	// Auto completion list
	g.complete = g.complete[:0]
	vars := g.console.Suggest(g.command, 5)
	for _, v := range vars {
		g.complete = append(g.complete, v.Name())
	}
	sort.Strings(g.complete)

	// If the backspace key is pressed, remove one character
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(g.command) >= 1 {
			t := []rune(g.command)
			g.command = string(t[:len(t)-1])
		}
	}

	// If the up is pressed, bring back previous command
	if repeatingKeyPressed(ebiten.KeyUp) {
		if g.prevCommand != "" {
			g.command = g.prevCommand
		}
	}

	// If the TAB is pressed get the first command from suggestions
	if repeatingKeyPressed(ebiten.KeyTab) {
		if len(g.complete) > 0 {
			g.command = g.complete[0]
		}
	}

	// If enter is pressed, process the command
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyKPEnter) {
		cv, err := g.console.ExecCmd(g.command)
		if err != nil {
			g.console.LogErrorf("%s -> %s", g.command, err)
		} else if !cv.IsFunc() {
			v, _ := cv.Interface()
			g.console.LogInfof("%s %v", cv.Name(), v)
		}
		g.prevCommand = g.command
		g.command = ""
	}

	g.counter++
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	cmdY := int(float64(456) * ebiten.DeviceScaleFactor())
	bufY := int(float64(440) * ebiten.DeviceScaleFactor())
	bufStr := g.console.BufferWrapped(90)
	size := text.BoundString(inconsolata.Regular8x16, bufStr)

	// Blink the cursor.
	t := "$ " + g.command
	if g.counter%60 < 30 {
		t += "_"
	}

	text.Draw(screen, bufStr, inconsolata.Regular8x16, 20, bufY-size.Max.Y, color.White)
	text.Draw(screen, t, inconsolata.Regular8x16, 20, cmdY, color.White)

	// Draw suggestion list
	if len(g.complete) > 0 {
		compStr := strings.Join(g.complete, "\n")
		text.Draw(screen, compStr, inconsolata.Regular8x16, 20, 40, color.RGBA{R: 255, G: 255, B: 0, A: 255})
	}
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	s := ebiten.DeviceScaleFactor()
	return int(float64(outsideWidth) * s), int(float64(outsideHeight) * s)
}

func main() {
	g := &game{
		console: convar.NewConsole(10, convar.LogError, "[INFO] ", "[WARNING] ", "[ERROR] "),
	}

	// Register variables
	g.console.RegDefaultConVars()
	g.console.RegConVar(convar.NewConVar(
		"cl_width", reflect.Int, false, "Sets client horizontal resolution", 640, func(con *convar.Console, oldVal, newVal interface{}) {
			con.LogInfof("Variable cl_width changed from %v to %v", oldVal, newVal)
			log.Printf("Variable cl_width changed from %v to %v", oldVal, newVal)
		},
	))
	g.console.RegConVar(convar.NewConVar(
		"cl_height", reflect.Int, false, "Sets client vertical resolution", 480, func(con *convar.Console, oldVal, newVal interface{}) {
			con.LogInfof("Variable cl_height changed from %v to %v", oldVal, newVal)
			log.Printf("Variable cl_height changed from %v to %v", oldVal, newVal)
		},
	))
	g.console.RegConVar(convar.NewConVar(
		"cl_title", reflect.String, false, "Sets window title", "convar test", func(con *convar.Console, oldVal, newVal interface{}) {
			con.LogInfof("Variable cl_title changed from %v to %v", oldVal, newVal)
			log.Printf("Variable cl_title changed from %v to %v", oldVal, newVal)
		},
	))

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Console Demo")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}
