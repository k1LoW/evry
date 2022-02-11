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
		want, err := execCmd(tt.cmd)
		if err != nil {
			t.Fatal(err, tt.cmd)
		}
		got, err := execCmd(fmt.Sprintf("%s | %s", tt.cmd, tt.evryCmd))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("\nwant %q\ngot  %q", want, got)
		}
	}
}

func TestMutex(t *testing.T) {
	cmd := `echo -e "2\n0\n1" | ./evry -l 1 -c 'xargs -I@ sh -c "sleep @; echo sleep @"'`
	want := "sleep 2\nsleep 0\nsleep 1\n"
	got, err := execCmd(cmd)
	if err != nil {
		t.Fatal(err, cmd)
	}
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func TestPipe(t *testing.T) {
	cmd := `echo -e "b\nc\na\ne\nd" | ./evry -l 10 -c 'cat | sort | head -3'`
	want := "a\nb\nc\n"
	got, err := execCmd(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func TestPipeWithArgs(t *testing.T) {
	cmd := `echo -e "b\nc\na\ne\nd" | ./evry -l 10 -- sh -c 'cat | sort | head -3'`
	want := "a\nb\nc\n"
	got, err := execCmd(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func execCmd(cmd string) (string, error) {
	b, err := exec.Command(os.Getenv("SHELL"), "-c", cmd).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
