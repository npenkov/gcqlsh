package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	Red     = color.New(color.FgRed).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	Blue    = color.New(color.FgBlue).SprintFunc()

	colors = map[color.Attribute]func(a ...interface{}) string{
		color.FgRed:     Red,
		color.FgMagenta: Magenta,
		color.FgGreen:   Green,
		color.FgYellow:  Yellow,
		color.FgBlue:    Blue,
	}
)

func PrintError(err string) {
	var addSpaceColor = 0
	if !color.NoColor {
		addSpaceColor = 9
	}
	fmt.Printf(fmt.Sprintf("%%%ds\n", 0+addSpaceColor), colors[color.FgRed](err))
}

func PrintColoredColumnVal(width int, val string, f func(a ...interface{}) string) {
	var addSpaceColor = 0
	if !color.NoColor {
		addSpaceColor = 9
	}
	fmt.Printf(fmt.Sprintf("| %%%ds ", width+addSpaceColor), f(val))
}

func PrintHeaderSeparator(width int) {
	fmt.Printf(fmt.Sprintf("+%%%ds", width+2), strings.Repeat("-", width+2))
}

func Colorful() bool {
	return !color.NoColor
}
