# cmdWorker

cmdWorker is the Golang library wrapping github.com/go-cmd/cmd adding timeouts,periodic output parsing using callback function and ability to abort command execution from output parsing callback function.

cmdWorker should not be used for console commands with huge output because go-cmd is used in buffering mode storing all stdout and stderr output in a memory

Usage:
```
cw := cmdWorker.NewCmdWorker(cmdPath, []string{}, readStdout)
status, err := cw.Run(nil)

if err != nil {
    if err == cmdWorker.ErrOperation {
      fmt.printf("%v finished with error %v", status.Exit)
    } else if err.(type) == *cmdWorker.InterruptedError  {
      fmt.printf("%v interrupted")
    }
}
```
