package cron

import (
	"container/heap"
	"sync"
	"time"
)

const defaultWaitTime = 1000000 * time.Hour

// Heap is minimal heap to implement Cron.
//
// It is safe for concurrent use by multiple goroutines.
type Heap struct {
	entries entries
	running bool

	add     chan *Entry
	remove  chan int
	stop    chan struct{}
	release chan struct{}
	lock    sync.Locker

	parser Parser
	lastID int
	logger Logger
}

// New return a Cron implement in min-heap.
//
// It is faster than simple array when adding tasks dynamically.
func New(opts ...Option) Cron {
	// TODO: localtime
	h := &Heap{
		add:     make(chan *Entry),
		remove:  make(chan int),
		stop:    make(chan struct{}),
		release: make(chan struct{}),
		lock:    NewSpinLock(),
		// lock: &sync.Mutex{},
		parser: NewParser(ParseOptionStandard),
		logger: defaultPrintLogger,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Add adds a job to the Cron to be run on the given schedule.
func (h *Heap) Add(spec string, job Job, opts ...EntryOption) int {
	schedule, err := h.parser.Parse(spec)
	if err != nil {
		h.logger.Error("Add job failure: %s", err)
		return 0
	}
	entry := &Entry{
		Spec:     spec,
		Schedule: schedule,
		Job:      job,
	}
	for _, opt := range opts {
		opt(entry)
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	h.lastID++
	id := h.lastID
	entry.ID = id
	if h.running {
		h.add <- entry
	} else {
		heap.Push(&h.entries, entry)
	}
	return id
}

// AddFunc adds a func to the Cron to be run on the given schedule.
func (h *Heap) AddFunc(spec string, fn func(), opts ...EntryOption) int {
	return h.Add(spec, FuncJob(fn), opts...)
}

// Remove an entry with entry-id.
func (h *Heap) Remove(id int) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.running {
		h.remove <- id
	} else {
		h.removeEntry(id)
	}
}

// Run the Cron in synchronous mode, or no-op if alreay running.
func (h *Heap) Run() {
	h.lock.Lock()
	if h.running {
		h.lock.Unlock()
		return
	}
	h.running = true
	h.lock.Unlock()
	h.logger.Info("Start cron")
	h.run()
}

func (h *Heap) run() {
	// Init all schedule
	now := time.Now()
	for _, e := range h.entries {
		e.Next = e.Schedule.Next(now)
		if e.RunFirst {
			go e.Job.Run()
			e.RunFirst = false
			e.count++
		}
	}
	// Init min-heap
	heap.Init(&h.entries)

	timer := time.NewTimer(0)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	for {
		// Reuse timer
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		if len(h.entries) == 0 || h.entries[0].Next.IsZero() {
			// Waiting along time util a new job join in.
			timer.Reset(defaultWaitTime)
		} else {
			timer.Reset(h.entries[0].Next.Sub(time.Now()))
		}

		for {
			select {
			case <-timer.C:
				if len(h.entries) == 0 {
					break
				}
				now = time.Now()
				// run job from heap
				for len(h.entries) > 0 {
					entry := h.entries[0]
					if entry.Next.After(now) || entry.Next.IsZero() {
						break
					}
					entry = heap.Pop(&h.entries).(*Entry)
					go entry.Job.Run()
					entry.count++
					if entry.Times != 0 && entry.count >= entry.Times {
						continue
					}
					entry.Prev = entry.Next
					entry.Next = entry.Schedule.Next(now)
					heap.Push(&h.entries, entry)
				}
			case entry := <-h.add:
				now = time.Now()
				entry.Next = entry.Schedule.Next(now)
				heap.Push(&h.entries, entry)
			case id := <-h.remove:
				h.removeEntry(id)
			case <-h.stop:
				return
			case <-h.release:
				return
			}
			break
		}
	}
}

// Stop the Cron if it is running, otherwise no-op.
func (h *Heap) Stop() {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.running {
		h.stop <- struct{}{}
		h.running = false
		h.logger.Info("Stop cron")
	}
}

// Release all entry in min-heap.
func (h *Heap) Release() {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.running {
		h.release <- struct{}{}
	}
	h.entries = h.entries[:0]
	h.logger.Info("Release cron")
}

func (h *Heap) removeEntry(id int) {
	for i, e := range h.entries {
		if e.ID == id {
			heap.Remove(&h.entries, i)
			// 循环里，只能删除一次，删除多了会导致 panic，命中了就删
			return
		}
	}
}

// entries implement container/heap
type entries []*Entry

func (e entries) Len() int      { return len(e) }
func (e entries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e entries) Less(i, j int) bool {
	// zero is greater than any other time
	if e[i].Next.IsZero() {
		return false
	}
	if e[j].Next.IsZero() {
		return true
	}
	return e[i].Next.Before(e[j].Next)
}

func (e *entries) Push(x interface{}) {
	if x == nil {
		return
	}
	*e = append(*e, x.(*Entry))
}

func (e *entries) Pop() interface{} {
	old := *e
	n := len(old)
	x := old[n-1]
	*e = old[0 : n-1]
	return x
}
