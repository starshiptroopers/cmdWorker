package cmdWorker

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestCmdWorker_Run(t *testing.T) {

	//parse output and write downloading progress
	readStdout := func(status OutputStatus) bool {
		progressRegexp := regexp.MustCompile(`^\d+\s+(\d+)`)
		//parse the line
		if len(status.newStderrLines) == 0 {
			return true
		}
		//remove leading \r
		l := strings.TrimPrefix(status.newStderrLines[len(status.newStderrLines)-1], "\r")
		matches := progressRegexp.FindStringSubmatchIndex(l)
		if len(matches) > 0 {
			s := string(progressRegexp.ExpandString(nil, "$1", l, matches))
			fmt.Printf("Downloaded %v bytes\n", s)
		}
		return true
	}
	//we download the www.google.com index page using curl with a speed limit of 2kb/s to demonstrate long time stdout parsing
	cw := NewCmdWorker("curl", []string{"www.google.com", "--limit-rate", "2000", "-o", "/dev/null"}, readStdout)
	status, err := cw.Run(nil)

	if err != nil {
		if err == ErrOperation {
			fmt.Printf("%v finished with error %v", status.Cmd, status.Exit)
		} else if _, ok := err.(*InterruptedError); ok {
			fmt.Printf("%v interrupted", status.Cmd)
		}
	}

	fmt.Printf("%v finished", status.Cmd)

}
