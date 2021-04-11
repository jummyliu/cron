package cron

import (
	"time"
)

// Job is an interface for submitted Cron jobs.
type Job interface {
	Run()
}

// FuncJob a func implement Job interface.
type FuncJob func()

// Run wrapper func
func (f FuncJob) Run() { f() }

// Entry the minimum task unit of Cron.
type Entry struct {
	ID       int
	Spec     string
	Job      Job
	Schedule Schedule
	Next     time.Time
	Prev     time.Time
	Times    uint // The max Execute times
	count    uint // already running count
}

type EntryOption func(e *Entry)

func WithEntryMaxExecuteTimes(times uint) EntryOption {
	return func(e *Entry) {
		e.Times = times
	}
}
