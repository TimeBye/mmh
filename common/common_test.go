package common

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		in   string
		cmd  string
		args []string
	}{
		{"ls", "ls", nil},
		{"ls -al /tmp", "ls", []string{"-al", "/tmp"}},
		{"  echo   hello  ", "echo", []string{"hello"}},
	}
	for _, c := range cases {
		cmd, args := ParseCommand(c.in)
		if cmd != c.cmd || !reflect.DeepEqual(args, c.args) {
			t.Errorf("ParseCommand(%q) = (%q, %v), want (%q, %v)", c.in, cmd, args, c.cmd, c.args)
		}
	}
}

func TestRenderedOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderedOutput(&buf, ColorLine{Prefix: "prod1", Value: "hello\n"}); err != nil {
		t.Fatalf("RenderedOutput returned error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "prod1") || !strings.Contains(out, "hello") {
		t.Fatalf("unexpected output: %q", out)
	}
}
