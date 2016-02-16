# Cron

http://crontab.org/

## Usage

```golang
package main

import (
        "fmt"

        "github.com/v2e4lisp/cron"
)

func main() {
        cron := cron.NewCron()
        hi := func() { fmt.Println("hello world") }
        // print a "hello world" every minute
        cron.Register("hi", "* * * * *", hi)
        cron.Start()
}
```

