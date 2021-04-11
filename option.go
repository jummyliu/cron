package cron

import "time"

type Option func(h *Heap)

func WithParser(p Parser) Option {
	return func(h *Heap) {
		h.parser = p
	}
}

func WithLogger(l Logger) Option {
	return func (h *Heap) {
		h.logger = l
	}
}

func WithLocaltime(t time.Time) Option {
	return func (h *Heap) {

	}
}
