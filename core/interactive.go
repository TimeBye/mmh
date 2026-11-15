package core

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/mritd/mmh/common"
	"golang.org/x/term"
)

const selectPageSize = 9

var selectActiveStyle = color.New(color.FgCyan, color.Bold)

type selectItem struct {
	cells []string
}

type listSelector struct {
	title   string
	headers []string
	items   []selectItem
	idx     int
	page    int
}

func (s *listSelector) pageCount() int {
	return (len(s.items) + selectPageSize - 1) / selectPageSize
}

func (s *listSelector) move(delta int) {
	s.idx += delta
	if s.idx < 0 {
		s.idx = 0
	}
	if s.idx >= len(s.items) {
		s.idx = len(s.items) - 1
	}
	s.page = s.idx / selectPageSize
}

func (s *listSelector) turnPage(delta int) {
	page := s.page + delta
	if page < 0 || page >= s.pageCount() {
		return
	}
	s.page = page
	s.idx = page * selectPageSize
}

func (s *listSelector) jump(n int) {
	idx := s.page*selectPageSize + n
	if idx < len(s.items) {
		s.idx = idx
	}
}

// render draws the selector and returns the number of lines written
func (s *listSelector) render(prevLines int) int {
	if prevLines > 0 {
		fmt.Printf("\x1b[%dF\x1b[0J", prevLines)
	}

	lines := s.buildLines()
	for _, line := range lines {
		fmt.Printf("%s\r\n", line)
	}
	return len(lines)
}

func (s *listSelector) buildLines() []string {
	widths := s.columnWidths()
	fitWidths(widths)

	lines := []string{fmt.Sprintf("%s (%d/%d):", s.title, s.page+1, s.pageCount())}
	lines = append(lines, border(widths, "┌", "┬", "┐"))
	lines = append(lines, s.renderRow(centerCells(s.headers, widths), widths, false))
	lines = append(lines, border(widths, "├", "┼", "┤"))

	start := s.page * selectPageSize
	end := start + selectPageSize
	if end > len(s.items) {
		end = len(s.items)
	}
	for i := start; i < end; i++ {
		marker := " "
		if i == s.idx {
			marker = "»"
		}
		cells := []string{marker, strconv.Itoa(i - start + 1)}
		cells = append(cells, s.items[i].cells...)
		lines = append(lines, s.renderRow(cells, widths, i == s.idx))
	}
	lines = append(lines, border(widths, "└", "┴", "┘"))
	lines = append(lines, "")
	lines = append(lines, "↑/↓ 选择  ←/→ 翻页  1-9 快速跳转  Enter 确认  q 退出")
	return lines
}

func (s *listSelector) columnWidths() []int {
	widths := make([]int, len(s.headers))
	for i, h := range s.headers {
		widths[i] = runewidth.StringWidth(h)
	}
	// the first column is the active marker, always one cell wide
	widths[0] = 1
	for _, item := range s.items {
		for i, c := range item.cells {
			if w := runewidth.StringWidth(c); w > widths[i+2] {
				widths[i+2] = w
			}
		}
	}
	return widths
}

func (s *listSelector) renderRow(cells []string, widths []int, active bool) string {
	var b strings.Builder
	b.WriteString("│")
	for i, c := range cells {
		text := runewidth.Truncate(c, widths[i], "…")
		b.WriteString(" ")
		b.WriteString(runewidth.FillRight(text, widths[i]))
		b.WriteString(" │")
	}
	row := b.String()
	if active {
		return selectActiveStyle.Sprint(row)
	}
	return row
}

func centerCells(cells []string, widths []int) []string {
	centered := make([]string, len(cells))
	for i, c := range cells {
		centered[i] = center(c, widths[i])
	}
	return centered
}

func center(s string, width int) string {
	gap := width - runewidth.StringWidth(s)
	if gap <= 0 {
		return s
	}
	left := gap / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", gap-left)
}

func border(widths []int, left, mid, right string) string {
	var b strings.Builder
	b.WriteString(left)
	for i, w := range widths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(widths)-1 {
			b.WriteString(mid)
		}
	}
	b.WriteString(right)
	return b.String()
}

// fitWidths shrinks the last column when the table is wider than the terminal
func fitWidths(widths []int) {
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termWidth <= 0 || len(widths) == 0 {
		return
	}
	total := 1
	for _, w := range widths {
		total += w + 3
	}
	if total <= termWidth {
		return
	}
	last := len(widths) - 1
	shrink := total - termWidth
	if widths[last]-shrink < 8 {
		widths[last] = 8
		return
	}
	widths[last] -= shrink
}

// clearLines erases the selector output
func clearLines(n int) {
	if n > 0 {
		fmt.Printf("\x1b[%dF\x1b[0J", n)
	}
}

// SelectServer displays the server list and returns the selected server name,
// it returns ok=false when canceled or when running in a non-interactive environment
func SelectServer() (string, bool) {
	return selectServer("Select Server")
}

// SelectConfig displays the config list and returns the selected config name
func SelectConfig() (string, bool) {
	if len(Configs) == 0 {
		fmt.Println("no configs found")
		return "", false
	}

	items := make([]selectItem, 0, len(Configs))
	for _, c := range Configs {
		name := c.Name
		if c.IsCurrent {
			name = "* " + name
		}
		items = append(items, selectItem{cells: []string{name, c.Path}})
	}

	idx, ok, err := runSelector("Select Config", []string{"", "#", "NAME", "PATH"}, items, ListConfigs)
	common.CheckAndExit(err)
	if !ok {
		return "", false
	}
	return Configs[idx].Name, true
}

// InteractiveLogin displays the server list and logs into the selected server
func InteractiveLogin() {
	name, ok := selectServer("Select Login Server")
	if !ok {
		return
	}
	SingleLogin(name)
}

func selectServer(title string) (string, bool) {
	servers := InteractiveServers()
	if len(servers) == 0 {
		fmt.Println("no servers found, please add servers to your config file")
		return "", false
	}

	hasTags := false
	for _, s := range servers {
		if len(s.Tags) > 0 {
			hasTags = true
			break
		}
	}

	headers := []string{"", "#", "NAME", "USER", "ADDRESS", "PORT"}
	if hasTags {
		headers = append(headers, "TAGS")
	}

	items := make([]selectItem, 0, len(servers))
	for _, s := range servers {
		cells := []string{s.Name, s.User, s.Address, strconv.Itoa(s.Port)}
		if hasTags {
			cells = append(cells, strings.Join(s.Tags, ","))
		}
		items = append(items, selectItem{cells: cells})
	}

	idx, ok, err := runSelector(title, headers, items, func() { PrintServers(true) })
	common.CheckAndExit(err)
	if !ok {
		return "", false
	}
	return servers[idx].Name, true
}

// runSelector runs the interactive list selector, it returns ok=false when canceled
func runSelector(title string, headers []string, items []selectItem, fallback func()) (int, bool, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) || !term.IsTerminal(int(os.Stdout.Fd())) {
		fallback()
		fmt.Printf("\n非交互环境无法交互选择，请直接指定参数\n")
		return 0, false, nil
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = term.Restore(fd, state) }()

	s := &listSelector{title: title, headers: headers, items: items}
	lines := s.render(0)
	r := bufio.NewReader(os.Stdin)
	for {
		b, err := r.ReadByte()
		if err != nil {
			clearLines(lines)
			return 0, false, err
		}
		switch b {
		case 0x03, 0x04, 'q', 'Q': // ctrl-c / ctrl-d / q
			clearLines(lines)
			return 0, false, nil
		case '\r', '\n':
			clearLines(lines)
			return s.idx, true, nil
		case 0x1b:
			seq, err := readEscapeSequence(r)
			if err != nil {
				clearLines(lines)
				return 0, false, err
			}
			switch seq {
			case "A":
				s.move(-1)
			case "B":
				s.move(1)
			case "D":
				s.turnPage(-1)
			case "C":
				s.turnPage(1)
			}
		default:
			if b >= '1' && b <= '9' {
				s.jump(int(b - '1'))
			}
		}
		lines = s.render(lines)
	}
}

// readEscapeSequence reads the final byte of a CSI/SS3 sequence
func readEscapeSequence(r *bufio.Reader) (string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	if b != '[' && b != 'O' {
		return string(b), nil
	}
	b, err = r.ReadByte()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
