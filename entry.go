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
	RunFirst bool
}

type EntryOption func(e *Entry)

// WithEntryMaxExecuteTimes set the max execute times of entry.
func WithEntryMaxExecuteTimes(times uint) EntryOption {
	return func(e *Entry) {
		e.Times = times
	}
}

// WithEntryRunFirst run job at first when the cron is running.
func WithEntryRunFirst() EntryOption {
	return func(e *Entry) {
		e.RunFirst = true
	}
}
