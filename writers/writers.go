package writers

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/stack-99/honeypot/writers/colors"
)

// Write print text with a cool "animation" like a hacker
func Write(w io.Writer, str string) {
	chars := strings.Split(str, "")

	for _, ch := range chars {
		fmt.Fprint(w, ch)
		time.Sleep(30 * time.Millisecond)
	}
}

// ColorWrite same as Write but colored
func ColorWrite(w io.Writer, str, color string) {
	Write(w, color+str+colors.Reset)
}

func WriteFast(w io.Writer, str string) {
	fmt.Fprint(w, str)
}

// ColorWrite same as Write but colored
func ColorWriteFast(w io.Writer, str, color string) {
	WriteFast(w, color+str+colors.Reset)
}

// PrintEnd print a newline char at w
func PrintEnd(w io.Writer, ends int) {
	for i := 0; i < ends; i++ {
		fmt.Fprint(w, "\n")
	}
}
