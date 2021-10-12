// Copyright 2021 The Starship Troopers Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// cmdWorker is the Golang library wrapping github.com/go-cmd/cmd adding timeouts,
// periodic output parsing using callback function and ability to abort command execution from output parsing callback function
// cmdWorker should not be used for console commands with huge output because go-cmd is used in buffering mode
// storing all stdout and stderr output in a memory

package cmdWorker

import (
	"errors"
	"io"
	"log"
	"strings"
	"time"
)
import gocmd "github.com/go-cmd/cmd"

type CmdWorker struct {
	cmd           *gocmd.Cmd
	args          []string
	timeout       time.Duration
	readTimeout   time.Duration
	readInterval  time.Duration
	Status        gocmd.Status
	outputHandler func(OutputStatus) (carryOn bool)
}

type OutputStatus struct {
	NewStdoutlines []string
	NewStderrLines []string
	gocmd.Status
}

var (
	DefTimeout      = time.Second * 120 //timeout the cmd must finished
	DefReadTimeout  = time.Second * 30  //timeout waiting new data from stdout/stderr
	DefReadInterval = time.Second * 1   //the interval outputHandler is called to process the output
)

const (
	InterruptedByTimer         = 1
	InterruptedByOutputHandler = 2
)

var (
	ErrOperation = errors.New("command finished with error")
)

type InterruptedError struct {
	Reason int
}

func (e *InterruptedError) Error() (str string) {
	str = "interrupted by signal"
	if e.Reason == InterruptedByTimer {
		str = "interrupted with timeout"
	} else if e.Reason == InterruptedByOutputHandler {
		str = "interrupted by output handler"
	}
	return
}

//NewCmdWorker creates new worker with cmd as console command, args as command args and outputHandler callback function
//outputHandler calls every readInterval and used to process a command output, can be nil
func NewCmdWorker(cmd string, args []string, outputHandler func(OutputStatus) bool) CmdWorker {
	w := CmdWorker{
		gocmd.NewCmd(cmd, args...),
		args,
		DefTimeout,
		DefReadTimeout,
		DefReadInterval,
		gocmd.Status{},
		outputHandler,
	}
	return w
}

//SetTimeout sets command's execution timeout
func (w *CmdWorker) SetTimeout(t time.Duration) {
	w.timeout = t
}

//SetReadTimeout sets command read timeout
func (w *CmdWorker) SetReadTimeout(t time.Duration) {
	w.readTimeout = t
}

//SetReadInterval sets command's ReadInterval, outputHandler callback is called on every ReadInterval tick
func (w *CmdWorker) SetReadInterval(t time.Duration) {
	w.readInterval = t
}

//SetOutputHandler sets command's output callback which is called on every ReadInterval tick
//outputHandler callback should return true to continue command execution, or false to abort command execution
func (w *CmdWorker) SetOutputHandler(outputHandler func(OutputStatus) bool) {
	w.outputHandler = outputHandler
}

//Run starts the command and blocked until command finished or timeout happened
//stdin is a io.Reader where the cmd input is reading from, can be nil
//status is gocmd.Status struct describing running command
//err is nil on success, ErrOperation when command is finished with exit status != 0
//or InterruptedError type when command was interrupted with timeout or by OutputHandler
func (w *CmdWorker) Run(stdin io.Reader) (status gocmd.Status, err error) {

	statusChan := w.cmd.StartWithStdin(stdin) // non-blocking
	ticker := time.NewTicker(w.readInterval)
	var iError error
	go func() {
		pStdoutLineCount := 0
		pStderrLineCount := 0
		lastOutput := time.Now()
		for range ticker.C {
			status := w.cmd.Status()
			stdoutLineCount := len(status.Stdout)
			stderrLineCount := len(status.Stderr)
			newStdoutLines := status.Stdout[pStdoutLineCount:]
			newStderrLines := status.Stderr[pStderrLineCount:]
			if pStdoutLineCount != stdoutLineCount {
				pStdoutLineCount = stdoutLineCount
				lastOutput = time.Now()
			}
			if pStderrLineCount != stderrLineCount {
				pStderrLineCount = stderrLineCount
				lastOutput = time.Now()
			}
			if time.Since(lastOutput) > w.readTimeout {
				if iError == nil {
					iError = &InterruptedError{InterruptedByTimer}
				}
				_ = w.cmd.Stop()
				continue
			}
			if w.outputHandler != nil {
				if !w.outputHandler(OutputStatus{newStdoutLines, newStderrLines, status}) {
					if iError == nil {
						iError = &InterruptedError{InterruptedByOutputHandler}
					}
					_ = w.cmd.Stop()
				}
			}
		}
	}()
	defer ticker.Stop()
	// Stop command after w.timeout
	t := time.NewTimer(w.timeout)
	go func() {
		<-t.C
		if iError == nil {
			iError = &InterruptedError{InterruptedByTimer}
		}
		_ = w.cmd.Stop()
	}()

	// Block waiting for command to exit, be stopped, or be killed
	w.Status = <-statusChan
	t.Stop()

	if !w.Status.Complete {
		log.Printf("failed command: %v %v", w.cmd.Name, strings.Join(w.cmd.Args, " "))
		if iError != nil {
			err = iError
		} else if w.Status.Error != nil {
			err = w.Status.Error
		} else {
			err = &InterruptedError{}
		}
	}

	if err == nil && w.Status.Exit != 0 {
		err = ErrOperation
	}

	return w.Status, err
}

//Stops aborts the command execution
func (w *CmdWorker) Stop() {
	_ = w.cmd.Stop()
}
