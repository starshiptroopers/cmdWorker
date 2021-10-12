# cmdWorker

cmdWorker is the Golang library wrapping github.com/go-cmd/cmd adding timeouts,periodic output parsing using callback function and ability to abort command execution from output parsing callback function.

cmdWorker should not be used for console commands with huge output because go-cmd is used in buffering mode storing all stdout and stderr output in a memory

Usage:
```
cw := cmdWorker.NewCmdWorker(cmdPath, []string{}, readStdout)
status, err := cw.Run(nil)

if err != nil {
    if err == cmdWorker.ErrOperation {
      fmt.Printf("%v finished with error %v", status.Cmd, status.Exit)
    } else if _, ok := err.(*cmdWorker.InterruptedError); ok {
      fmt.printf("%v interrupted")
    }
}
```

Another yet example:
Start file downloading using curl with downloading progress processing

```
readStdout := func(status OutputStatus) bool {
        //regexp to parse the curl progress line (we need the second number)
		progressRegexp := regexp.MustCompile(`^\d+\s+(\d+)`)
		
		if len(status.newStderrLines) == 0 {
			return true
		}
		//get the last line from the output removing leading \r
		l := strings.TrimPrefix(status.newStderrLines[len(status.newStderrLines)-1], "\r")
		matches := progressRegexp.FindStringSubmatchIndex(l)
		if len(matches) > 0 {
			s := string(progressRegexp.ExpandString(nil, "$1", l, matches))
			fmt.Printf("Downloaded %v bytes\n", s)
		}
		return true
	}

//download the www.google.com index page using curl with a speed limit of 2kb/s to demonstrate long time stdout parsing
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
```

