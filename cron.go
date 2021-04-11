package cron

// Cron interface
type Cron interface {
	// Add adds a job to the Cron to be run on the given schedule.
	Add(spec string, job Job, opts ...EntryOption) int
	// AddFunc adds a func to the Cron to be run on the given schedule.
	AddFunc(spec string, fn func(), opts ...EntryOption) int
	// Remove an entry with entry-id.
	Remove(id int)
	// Run the Cron in synchronous mode, or no-op if alreay running.
	Run()
	// Stop the Cron if it is running, otherwise no-op.
	Stop()
	// Release all entry in Cron.
	Release()
}
