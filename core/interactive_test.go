package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestInteractiveServers(t *testing.T) {
	oldBasic, oldCurrent := basicConfig, currentConfig
	defer func() { basicConfig, currentConfig = oldBasic, oldCurrent }()

	basicConfig = Config{configPath: "basic.yaml", Servers: Servers{&Server{Name: "basic-srv"}}}
	currentConfig = Config{configPath: "prod.yaml", Servers: Servers{&Server{Name: "z-srv"}, &Server{Name: "a-srv"}}}

	got := InteractiveServers()
	if len(got) != 2 || got[0].Name != "z-srv" || got[1].Name != "a-srv" {
		t.Fatalf("non-basic context should keep YAML order, got %d", len(got))
	}

	currentConfig = basicConfig
	got = InteractiveServers()
	if len(got) != 1 || got[0].Name != "basic-srv" {
		t.Fatalf("basic context should contain basic servers, got %d", len(got))
	}
}

func TestSelectorTableWidth(t *testing.T) {
	s := &listSelector{
		title:   "Select Config",
		headers: []string{"", "#", "NAME", "PATH"},
		items: []selectItem{
			{cells: []string{"001-hand@基础架构", "/tmp/a.yaml"}},
			{cells: []string{"Skin79", "/tmp/b.yaml"}},
			{cells: []string{"* 中水", "/tmp/c.yaml"}},
		},
	}

	lines := s.buildLines()
	var tableLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "┌") || strings.HasPrefix(line, "│") || strings.HasPrefix(line, "├") || strings.HasPrefix(line, "└") {
			tableLines = append(tableLines, line)
		}
	}
	if len(tableLines) < 5 {
		t.Fatalf("unexpected table lines: %d", len(tableLines))
	}

	want := tableWidth(tableLines[0])
	for i, line := range tableLines[1:] {
		if got := tableWidth(line); got != want {
			t.Fatalf("line %d width = %d, want %d: %q", i+1, got, want, line)
		}
	}
}

// tableWidth counts border characters as one cell, CJK characters as two
func tableWidth(s string) int {
	width := 0
	for _, r := range s {
		if strings.ContainsRune("┌┬┐├┼┤└┴┘│─", r) {
			width++
			continue
		}
		width += runewidth.RuneWidth(r)
	}
	return width
}

func TestServerSelector(t *testing.T) {
	items := make([]selectItem, 12)
	for i := range items {
		items[i] = selectItem{cells: []string{fmt.Sprintf("s%d", i)}}
	}
	s := &listSelector{items: items}

	if got := s.pageCount(); got != 2 {
		t.Fatalf("pageCount = %d, want 2", got)
	}

	s.move(-1)
	if s.idx != 0 || s.page != 0 {
		t.Fatalf("move up from first: idx=%d page=%d", s.idx, s.page)
	}

	s.move(10)
	if s.idx != 10 || s.page != 1 {
		t.Fatalf("move down 10: idx=%d page=%d", s.idx, s.page)
	}

	s.move(10)
	if s.idx != 11 || s.page != 1 {
		t.Fatalf("move past end: idx=%d page=%d", s.idx, s.page)
	}

	s.jump(0)
	if s.idx != 9 {
		t.Fatalf("jump on page 1: idx=%d, want 9", s.idx)
	}

	s.turnPage(-1)
	if s.idx != 0 || s.page != 0 {
		t.Fatalf("turn page left: idx=%d page=%d", s.idx, s.page)
	}

	s.turnPage(-1)
	if s.idx != 0 || s.page != 0 {
		t.Fatalf("turn page left at first page: idx=%d page=%d", s.idx, s.page)
	}

	s.turnPage(1)
	if s.idx != 9 || s.page != 1 {
		t.Fatalf("turn page right: idx=%d page=%d", s.idx, s.page)
	}

	s.turnPage(1)
	if s.idx != 9 || s.page != 1 {
		t.Fatalf("turn page right at last page: idx=%d page=%d", s.idx, s.page)
	}
}
