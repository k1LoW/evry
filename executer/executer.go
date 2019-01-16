package executer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
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
	command []string
	ctx     context.Context
	cancel  context.CancelFunc
	timeout int
	out     chan *Output
}

// NewExecuter ...
func NewExecuter(ctx context.Context, command []string, timeout int) (*Executer, error) {
	innerCtx, cancel := context.WithCancel(ctx)

	return &Executer{
		command: command,
		ctx:     innerCtx,
		cancel:  cancel,
		timeout: timeout,
		out:     make(chan *Output, 1000),
	}, nil
}

// NewOutput ...
func (e *Executer) NewOutput() *Output {
	out := NewOutput()
	e.out <- out
	return out
}

// Execute ...
func (e *Executer) Execute(out *Output, in *bytes.Buffer) {
	defer func() {
		in.Reset()
		out.Unlock()
	}()

	innerCtx, cancel := context.WithTimeout(e.ctx, time.Duration(e.timeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var err error

	cmd := exec.CommandContext(innerCtx, e.command[0], e.command[1:]...)
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	cmd.Stdout = &stdout

	if err = cmd.Start(); err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	_, err = stdin.Write(in.Bytes()) // FIXME
	if err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	err = stdin.Close()
	if err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	if err = cmd.Wait(); err != nil {
		out.Stderr = fmt.Sprintf("%s", err)
		return
	}

	out.Stdout = stdout.String()
}

// Out ...
func (e *Executer) Out() <-chan *Output {
	return e.out
}
