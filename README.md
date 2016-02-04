daemongo
===

``daemongo`` is a simple Go(golang) daemon package.

## Installation

```
$ go get github.com/Iwark/daemongo
```

## Example

```go
package main

import (
  "flag"
  "log"

  "github.com/Iwark/daemongo"
)

var (
  child   = flag.Bool("child", false, "Run as a child process")
  logfile = flag.String("l", "logfile.log", "log file")
)

func main() {

  flag.Parse()
  if err := daemongo.Start(*child, *logfile); err != nil {
    log.Fatal(err)
    return
  }

  // anything to do ...
}
```