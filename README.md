godaemon
===

``godaemon`` is a simple Go(golang) daemon package.

## Installation

```
$ go get -u github.com/Iwark/godaemon
```

## Example

```go
package main

import (
  "flag"
  "log"

  "github.com/Iwark/godaemon"
)

var (
  child   = flag.Bool("child", false, "Run as a child process")
  logfile = flag.String("l", "logfile.log", "log file")
)

func main() {

  flag.Parse()
  if err := godaemon.Start(*child); err != nil {
    log.Fatal(err)
    return
  }
  f, err := godaemon.OutputFile(*logfile)
  if err != nil {
    log.Fatal(err)
    return
  }
  log.SetOutput(f)
  f.Close()

  // anything to do ...
}
```