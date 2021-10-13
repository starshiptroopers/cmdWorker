# cmdWorker

cmdWorker is the Golang library wrapping github.com/go-cmd/cmd adding timeouts and callback to parse stdout and stderr at regular intervals

It should not be used for console commands with huge output because go-cmd is used in buffering mode storing all stdout and stderr output in a memory

Three configurable timeouts is supoorted:

```
var (
	DefTimeout      = time.Second * 120 //timeout the cmd must finished
	DefReadTimeout  = time.Second * 30  //timeout waiting new data from stdout/stderr
	DefReadInterval = time.Second * 1   //the interval outputHandler is called to process the output
)
```

##### Usage:
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

##### Another yet example:
Do a file downloading using curl with progress processing

```
readStdout := func(status cmdWorker.OutputStatus) bool {
        //regexp to parse the curl progress line (we need the second number)
		progressRegexp := regexp.MustCompile(`^\d+\s+(\d+)`)
		
		if len(status.NewStderrLines) == 0 {
			return true
		}
		//get the last line from the output removing leading \r
		l := strings.TrimPrefix(status.NewStderrLines[len(status.NewStderrLines)-1], "\r")
		matches := progressRegexp.FindStringSubmatchIndex(l)
		if len(matches) > 0 {
			s := string(progressRegexp.ExpandString(nil, "$1", l, matches))
			fmt.Printf("Downloaded %v bytes\n", s)
		}
		return true
	}

//download the www.google.com index page using curl with a speed limit of 2kb/s to demonstrate long time stdout parsing
cw := cmdWorker.NewCmdWorker("curl", []string{"www.google.com", "--limit-rate", "2000", "-o", "/dev/null"}, readStdout)
status, err := cw.Run(nil)

if err != nil {
    if err == cmdWorker.ErrOperation {
        fmt.Printf("%v finished with error %v", status.Cmd, status.Exit)
    } else if _, ok := err.(*InterruptedError); ok {
        fmt.Printf("%v interrupted", status.Cmd)
    }
}

fmt.Printf("%v finished", status.Cmd)
```

