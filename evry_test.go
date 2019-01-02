package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var tests = []struct {
	cmd     string
	evryCmd string
}{
	{"echo evry", "./evry -l 1"},
	{"echo evry", "./evry -l 3"},
	{`echo -e "a\nb\nc"`, "./evry -l 3"},
	{"cat LICENSE", "./evry -l 1"},
	{"cat LICENSE", "./evry -l 3"},
	{"cat LICENSE", "./evry -l 1 | cat"},
	{"cat LICENSE", "./evry -l 3 | cat"},
	{"echo evry", "./evry -s 1"},
	{"echo evry", "./evry -s 3"},
	{`echo -e "a\nb\nc"`, "./evry -s 3"},
	{"cat LICENSE", "./evry -s 1"},
	{"cat LICENSE", "./evry -s 3"},
	{"cat LICENSE", "./evry -s 1 | cat"},
	{"cat LICENSE", "./evry -s 3 | cat"},
}

func TestCat(t *testing.T) {
	for _, tt := range tests {
		want := execCmd(tt.cmd)
		got := execCmd(fmt.Sprintf("%s | %s", tt.cmd, tt.evryCmd))
		if got != want {
			t.Errorf("\nwant %q\ngot  %q", want, got)
		}
	}
}

func TestMutex(t *testing.T) {
	cmd := `echo -e "2\n0\n1" | ./evry -l 1 -c 'xargs -I@ sh -c "sleep @; echo sleep @"'`
	want := "sleep 2\nsleep 0\nsleep 1\n"
	got := execCmd(cmd)
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func TestPipe(t *testing.T) {
	cmd := `echo -e "b\nc\na\ne\nd" | ./evry -l 10 -c 'cat | sort | head -3'`
	want := "a\nb\nc\n"
	got := execCmd(cmd)
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func execCmd(cmd string) string {
	b, _ := exec.Command(os.Getenv("SHELL"), "-c", cmd).Output()
	return string(b)
}
