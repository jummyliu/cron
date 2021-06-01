# Introduction
Library `cron` implements a timewheel in min-heap. It is a time-based job scheduler.

It requires Go 1.11 or later due to usage of Go Modules.

# Feature
The default cron expression support 5 fields. (Minutes Hours DayOfMonth Month DayOfWeek)

You can use custom expressions to support seconds.
e.g.
```go
// Seconds Minutes Hours DayOfMonth Month DayOfWeek
c := cron.New(cron.WithParser(cron.NewParser(cron.ParseOptionAll)))
```

- Limit job execution times.
```go
c.AddFunc("* * * * *", func() {
	fmt.Println("Every minite, and the max execute times is 2.")
}, cron.WithEntryMaxExecuteTimes(2))
```

- Run job at first when the cron is running.
```go
c.AddFunc(
	"0 0 1 1 *",
	func() { fmt.Println("Run first at cron is running.")},
	cron.WithEntryRunFirst(),
)
```

# How to install
```bash
go get -u github.com/jummyliu/cron
```

# How to use

```go
package main

import (
	"fmt"
	"time"

	"github.com/jummyliu/cron"
)

func main() {
	// The default cron does not support seconds (Minutes Hours DayOfMonth Month DayOfWeek)
	c := cron.New()

	// Add job before start
	c.Add("* * * * *", cron.FuncJob(func() {
		fmt.Println("Every minite, and the max execute times is 2.")
	}), cron.WithEntryMaxExecuteTimes(2))
	c.AddFunc(
		"0 0 1 1 *",
		func() { fmt.Println("Run it at first when the cron is running.")},
		cron.WithEntryRunFirst(),
	)
	id := c.AddFunc(cron.DescriptorEveryPrefix + "2s", func() {
		fmt.Println("Every 2 seconds")
	})

	// Start cron
	go c.Run()

	// Remove task by id
	c.Remove(id)

	// Add job after start
	c.Add("* * * * *", cron.FuncJob(func() {
		fmt.Println("Every minite")
	}))

	c.AddFunc(cron.DescriptorEveryPrefix + "2s", func() {
		fmt.Println("Every 2 seconds")
	})

	time.Sleep(time.Second * 10)

	// Stop cron
	c.Stop()
	// Remove all tasks
	c.Release()
}
```

# Cron expression format

There are two cron spec formats in common usage:

- The "standard" cron format, described on [the Cron wikipedia page](https://en.wikipedia.org/wiki/Cron) ( [zh](https://zh.wikipedia.org/wiki/Cron) ) commonly used by the cron Linux system utility.
- The cron format used by [the Quartz Scheduler](http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/tutorial-lesson-06.html), common used for scheduled jobs in Java software.

## A cron expression represents a set of times, using 6 space-separated fields.

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

### Special Characters

- Asterisk ( * )

The asterisk indicates that the cron expression will match for all values of the
field; e.g., using an asterisk in the 5th field (month) would indicate every
month.

- Slash ( / )

Slashes are used to describe increments of ranges. For example 3-59/15 in the
1st field (minutes) would indicate the 3rd minute of the hour and every 15
minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...",
that is, an increment over the largest possible range of the field.  The form
"N/..." is accepted as meaning "N-MAX/...", that is, starting at N, use the
increment until the end of that specific range.  It does not wrap around.

- Comma ( , )

Commas are used to separate items of a list. For example, using "MON,WED,FRI" in 
the 5th field (day of week) means Mondays, Wednesdays and Fridays.

- Dash ( - )

Dash defines ranges. For example, 10–30 indicates every minute between 10 and 30, inclusive.

- Question mark ( ? )

Question mark may be used instead of '*' for leaving either day-of-month or
day-of-week blank.

## You may use one of several pre-defined schedules in place of a cron expression.

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

where "duration" is a string accepted by [time.ParseDuration](https://pkg.go.dev/time#ParseDuration).

# FAQ
1. If the dayOfWeek field and dayOfMonth field are not both equal to '*' or '?', the two fields are logical or relational, otherwise they are logical and relational.

# Why use min-heap instead of array?
Add tasks dynamically in min-heap is much faster than array when Cron is running.

## In benchmark:
It is less than 1 seconds to add 100,000 jobs.

But in `github.com/robfig/cron/v3`, it takes more than 2 minutes to do the same thing.

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
	// c.Stop()
	for i := 0; i < 100000; i++ {
		// c.AddFunc("*/2 * * * * *", func() {
		// 	atomic.AddInt64(&count, 1)
		// })
		c.AddFunc("0 0 0 1 1 *", func() {
			atomic.AddInt64(&count, 1)
		})
	}
	finishToAdd := time.Now()
	time.Sleep(10 * time.Second)

	c.Stop()
	end := time.Now()
	fmt.Println("Start cron            :", begin)
	fmt.Println("Finish to add jobs    :", finishToAdd)
	fmt.Println("Stop cron             :", end)
	fmt.Println("count", count)

	fmt.Printf("\nIt tasks %v to add jobs\n", finishToAdd.Sub(begin))
}
```

output
```sh
Start cron            : 2021-06-01 15:20:34.447978 +0800 CST m=+0.009001201
Finish to add jobs    : 2021-06-01 15:20:35.336744 +0800 CST m=+0.897772901
Stop cron             : 2021-06-01 15:20:45.3436333 +0800 CST m=+10.904727301
count 0

It tasks 888.7717ms to add jobs
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
	// c.Stop()
	for i := 0; i < 100000; i++ {
		// c.AddFunc("*/2 * * * * *", func() {
		// 	atomic.AddInt64(&count, 1)
		// })
		c.AddFunc("0 0 0 1 1 *", func() {
			atomic.AddInt64(&count, 1)
		})
	}
	finishToAdd := time.Now()
	time.Sleep(10 * time.Second)

	c.Stop()
	end := time.Now()
	fmt.Println("Start cron            :", begin)
	fmt.Println("Finish to add jobs    :", finishToAdd)
	fmt.Println("Stop cron             :", end)
	fmt.Println("count", count)

	fmt.Printf("\nIt tasks %v to add jobs\n", finishToAdd.Sub(begin))
}
```

output
```sh
Start cron            : 2021-06-01 15:17:35.0382169 +0800 CST m=+0.004945601
Finish to add jobs    : 2021-06-01 15:19:58.1361499 +0800 CST m=+143.103845201
Stop cron             : 2021-06-01 15:20:08.1506088 +0800 CST m=+153.118369201
count 0

It tasks 2m23.0988996s to add jobs
```
