package executer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	shellwords "github.com/mattn/go-shellwords"
)

// Output ...
type Output struct {
	Stdout string
	Stderr string
	mutex  *sync.Mutex
}

// NewOutput ...
func NewOutput() *Output {
	out := &Output{
		mutex: new(sync.Mutex),
	}
	out.Lock()
	return out
}

// Lock ...
func (o *Output) Lock() {
	o.mutex.Lock()
}

// Unlock ...
func (o *Output) Unlock() {
	o.mutex.Unlock()
}

// OutputCombime ...
func (o *Output) OutputCombime() {
	o.Lock()
	if o.Stdout != "" {
		fmt.Printf(o.Stdout)
	}
	if o.Stderr != "" {
		_, _ = fmt.Fprintf(os.Stderr, "%s", o.Stderr)
	}
	o.Unlock()
}

// Executer ...
type Executer struct {
	commands [][]string
	ctx      context.Context
	cancel   context.CancelFunc
	timeout  int
	out      chan *Output
}

// NewExecuter ...
func NewExecuter(ctx context.Context, command string, timeout int) (*Executer, error) {
	innerCtx, cancel := context.WithCancel(ctx)
	commands := [][]string{}
	parser := shellwords.NewParser()
	for {
		c, err := parser.Parse(command)
		if err != nil {
			cancel()
			return nil, err
		}
		commands = append(commands, c)
		pos := parser.Position
		if pos < 0 {
			break
		}
		if string(command[pos]) != "|" {
			break
		}
		command = command[pos+1:]
	}

	return &Executer{
		commands: commands,
		ctx:      innerCtx,
		cancel:   cancel,
		timeout:  timeout,
		out:      make(chan *Output, 1000),
	}, nil
}

// NewOutput ...
func (e *Executer) NewOutput() *Output {
	out := NewOutput()
	e.out <- out
	return out
}

// Execute ...
func (e *Executer) Execute(out *Output, in []byte) {
	defer func() {
		out.Unlock()
	}()

	innerCtx, cancel := context.WithTimeout(e.ctx, time.Duration(e.timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// reference: https://github.com/mattn/go-pipeline/blob/master/pipeline.go#L9
	cmds := make([]*exec.Cmd, len(e.commands))
	var err error

	for i, c := range e.commands {
		cmds[i] = exec.CommandContext(innerCtx, c[0], c[1:]...)
		if i > 0 {
			if cmds[i].Stdin, err = cmds[i-1].StdoutPipe(); err != nil {
				out.Stderr = fmt.Sprintf("%s", err)
				return
			}
		}
		cmds[i].Stderr = &stderr
	}

	stdin, err := cmds[0].StdinPipe()
	if err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	cmds[len(cmds)-1].Stdout = &stdout

	for _, c := range cmds {
		if err = c.Start(); err != nil {
			out.Stderr = fmt.Sprintf("%s", err)
			return
		}
	}

	_, err = stdin.Write(in)
	if err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	err = stdin.Close()
	if err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	for _, c := range cmds {
		if err = c.Wait(); err != nil {
			out.Stderr = fmt.Sprintf("%s", err)
			return
		}
	}

	out.Stdout = stdout.String()
}

// Out ...
func (e *Executer) Out() <-chan *Output {
	return e.out
}
