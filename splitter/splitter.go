package splitter

import (
	"context"
	"sync"
	"time"

	"github.com/k1LoW/evry/executer"
)

// Splitter ...
type Splitter interface {
	Start()
	Stop()
	In([]byte)
	Close()
	Done() <-chan struct{}
}

// LineSplitter ...
type LineSplitter struct {
	interval int
	in       chan []byte
	executer *executer.Executer
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewLineSplitter ...
func NewLineSplitter(ctx context.Context, line int, command string, timeout int) (*LineSplitter, error) {
	innerCtx, cancel := context.WithCancel(ctx)

	e, err := executer.NewExecuter(innerCtx, command, timeout)
	if err != nil {
		cancel()
		return nil, err
	}
	return &LineSplitter{
		interval: line,
		in:       make(chan []byte, 10000),
		executer: e,
		ctx:      innerCtx,
		cancel:   cancel,
	}, nil
}

// Start ...
func (s *LineSplitter) Start() {
	defer s.Stop()

	count := 0
	buffer := []byte{}
	wg := &sync.WaitGroup{}

	// output
	go func() {
		for out := range s.executer.Out() {
			out.OutputCombime()
			wg.Done()
		}
	}()

L:
	for {
		select {
		case in := <-s.in:
			if in == nil {
				if len(buffer) > 0 {
					dst := make([]byte, len(buffer))
					copy(dst, buffer)
					wg.Add(1)
					out := s.executer.NewOutput()
					go s.executer.Execute(out, dst)
					buffer = nil
					count = 0
				}
				break L
			}
			buffer = append(buffer, in...)
			count++
			if count == s.interval {
				dst := make([]byte, len(buffer))
				copy(dst, buffer)
				wg.Add(1)
				out := s.executer.NewOutput()
				go s.executer.Execute(out, dst)
				buffer = nil
				count = 0
			}
		case <-s.ctx.Done():
			break L
		}
	}
	wg.Wait()
}

// Stop ...
func (s *LineSplitter) Stop() {
	s.cancel()
}

// In ...
func (s *LineSplitter) In(in []byte) {
	s.in <- in
}

// Close ...
func (s *LineSplitter) Close() {
	close(s.in)
}

// Done ...
func (s *LineSplitter) Done() <-chan struct{} {
	return s.ctx.Done()
}

// SecSplitter ...
type SecSplitter struct {
	interval int
	in       chan []byte
	executer *executer.Executer
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSecSplitter ...
func NewSecSplitter(ctx context.Context, sec int, command string, timeout int) (*SecSplitter, error) {
	innerCtx, cancel := context.WithCancel(ctx)

	e, err := executer.NewExecuter(innerCtx, command, timeout)
	if err != nil {
		cancel()
		return nil, err
	}
	return &SecSplitter{
		interval: sec,
		in:       make(chan []byte, 10000),
		executer: e,
		ctx:      innerCtx,
		cancel:   cancel,
	}, nil
}

// Start ...
func (s *SecSplitter) Start() {
	defer s.Stop()
	eol := false

	ticker := time.NewTicker(time.Duration(s.interval) * time.Second)
	buffer := []byte{}
	wg := &sync.WaitGroup{}

	// output
	go func() {
		for out := range s.executer.Out() {
			out.OutputCombime()
			wg.Done()
		}
	}()

L:
	for {
		select {
		case in := <-s.in:
			if in == nil {
				eol = true
			} else {
				buffer = append(buffer, in...)
			}
		case <-ticker.C:
			dst := make([]byte, len(buffer))
			copy(dst, buffer)
			wg.Add(1)
			out := s.executer.NewOutput()
			go s.executer.Execute(out, dst)
			buffer = nil
			if eol {
				break L
			}
		case <-s.ctx.Done():
			break L
		}
	}
	wg.Wait()
}

// Stop ...
func (s *SecSplitter) Stop() {
	s.cancel()
}

// In ...
func (s *SecSplitter) In(in []byte) {
	s.in <- in
}

// Close ...
func (s *SecSplitter) Close() {
	close(s.in)
}

// Done ...
func (s *SecSplitter) Done() <-chan struct{} {
	return s.ctx.Done()
}
