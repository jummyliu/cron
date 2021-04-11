# Introduction
Library `cron` implements a timewheel in min-heap.

It requires Go 1.11 or later due to usage of Go Modules.

# Features

# How to install
```bash
go get -u github.com/jummyliu/cron
```

# How to use

```go
package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jummyliu/cron"
)

func main() {
	c := cron.New(cron.WithParser(cron.NewParser(cron.ParseOptionAll)))
	var count int64 = 0

    // Add job before start
	c.Add("* * * * * *", cron.FuncJob(func() {
		atomic.AddInt64(&count, 1)
	}), cron.WithEntryMaxExecuteTimes(2)) // The max execute times is 2.

	c.AddFunc(cron.DescriptorEveryPrefix + "2s", func() {
		atomic.AddInt64(&count, 1)
	})

    // Start cron
	go c.Run()

    // Add job after start
	c.Add("* * * * * *", cron.FuncJob(func() {
		atomic.AddInt64(&count, 1)
	}))

	c.AddFunc(cron.DescriptorEveryPrefix + "2s", func() {
		atomic.AddInt64(&count, 1)
	})

	c.Stop()
}
```

A cron expression represents a set of times, using 6 space-separated fields.

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?


You may use one of several pre-defined schedules in place of a cron expression.

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 * * * *

Intervals

You may also schedule a job to execute at fixed intervals, starting at the time it's added
or cron is run. This is supported by formatting the cron spec like this:

    @every <duration>

where "duration" is a string accepted by time.ParseDuration
(http://golang.org/pkg/time/#ParseDuration).

# why use min-heap instead of array?
Add tasks dynamically in min-heap is much faster than array when Cron is running.

## In benchmark:
It is less than 1 seconds to add 100,000 jobs.

But in `github.com/robfig/cron/v3`, it takes about three minutes to do the same thing.

### `github.com/jummyliu/cron`
```go
package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jummyliu/cron"
)

func main() {
	c := cron.New(cron.WithParser(cron.NewParser(cron.ParseOptionAll)))
	var count int64 = 0
	go c.Run()
	begin := time.Now()
	fmt.Println("Start cron", begin)
	// c.Stop()
	for i := 0; i < 100000; i++ {
		c.AddFunc("*/2 * * * * *", func() {
			atomic.AddInt64(&count, 1)
		})
	}
	finishToAdd := time.Now()
	fmt.Println("Finish to add jobs", finishToAdd)
	time.Sleep(10 * time.Second)

	end := time.Now()
	fmt.Println("count", count)
	fmt.Println("end time", end)

	fmt.Printf("\nIt tasks %v to add jobs\n", finishToAdd.Sub(begin))
}
```

output
```sh
Start cron 2021-04-11 14:08:16.8818482 +0800 CST m=+0.005998701
Finish to add jobs 2021-04-11 14:08:17.4918481 +0800 CST m=+0.615998601
count 500000
end time 2021-04-11 14:08:27.4978492 +0800 CST m=+10.621999701

It tasks 609.9999ms to add jobs
```

### `github.com/robfig/cron/v3`
```go
package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	c := cron.New(cron.WithSeconds())
	var count int64 = 0
	go c.Run()
	begin := time.Now()
	fmt.Println("Start cron", begin)
	// c.Stop()
	for i := 0; i < 100000; i++ {
		c.AddFunc("*/2 * * * * *", func() {
			atomic.AddInt64(&count, 1)
		})
	}
	finishToAdd := time.Now()
	fmt.Println("Finish to add jobs", finishToAdd)
	time.Sleep(10 * time.Second)

	end := time.Now()
	fmt.Println("count", count)
	fmt.Println("end time", end)

	fmt.Printf("\nIt tasks %v to add jobs\n", finishToAdd.Sub(begin))
}
```

output
```sh
Start cron 2021-04-11 14:09:30.5654322 +0800 CST m=+0.006000201
Finish to add jobs 2021-04-11 14:12:40.4779312 +0800 CST m=+189.918499201
count 1464874
end time 2021-04-11 14:12:50.4833052 +0800 CST m=+199.923873201

It tasks 3m9.912499s to add jobs
```