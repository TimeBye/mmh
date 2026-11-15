package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"
	"text/template"

	"github.com/fatih/color"
)

const renderTpl = `{{ %s .Prefix }} {{ .Value }}`

const (
	ColorRed     = "red"
	ColorGreen   = "green"
	ColorYellow  = "yellow"
	ColorBlue    = "blue"
	ColorMagenta = "magenta"
	ColorCyan    = "cyan"
	ColorWhite   = "white"
)

// colorOrder is the rotation order used to assign colors to output prefixes
var colorOrder = []string{
	ColorRed,
	ColorGreen,
	ColorYellow,
	ColorBlue,
	ColorMagenta,
	ColorCyan,
	ColorWhite,
}

var colorFuncMap = template.FuncMap{
	ColorRed:     color.New(color.FgRed).SprintfFunc(),
	ColorGreen:   color.New(color.FgGreen).SprintfFunc(),
	ColorYellow:  color.New(color.FgYellow).SprintfFunc(),
	ColorBlue:    color.New(color.FgBlue).SprintfFunc(),
	ColorMagenta: color.New(color.FgMagenta).SprintfFunc(),
	ColorCyan:    color.New(color.FgCyan).SprintfFunc(),
	ColorWhite:   color.New(color.FgWhite).SprintfFunc(),
}

var (
	tplCacheMux sync.Mutex
	tplCacheMap = make(map[string]*template.Template)
	colorIndex  int
)

type ColorLine struct {
	Prefix string
	Value  string
}

// RenderedOutput writes a line with a colored prefix, every distinct prefix is
// assigned a color in rotation and its template is cached for reuse
func RenderedOutput(wr io.Writer, line ColorLine) error {
	tplCacheMux.Lock()
	tpl, ok := tplCacheMap[line.Prefix]
	if !ok {
		colorName := colorOrder[colorIndex%len(colorOrder)]
		colorIndex++
		tpl = template.Must(template.New("").Funcs(colorFuncMap).Parse(fmt.Sprintf(renderTpl, colorName)))
		tplCacheMap[line.Prefix] = tpl
	}
	tplCacheMux.Unlock()
	return tpl.Execute(wr, line)
}

func Converted2Rendered(r io.Reader, w io.Writer, prefix string) {
	reader := bufio.NewReader(r)
	// use buf to ensure atomic output of each line
	var buf bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		_ = RenderedOutput(&buf, ColorLine{Prefix: prefix, Value: line})
		_, _ = io.Copy(w, &buf)
		buf.Reset()
	}
}
