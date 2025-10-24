package funtimer

import "time"

type ChannelTimer struct {
	C chan time.Time
}

func (t *ChannelTimer) Stop() bool {}

func NewTimer(d time.Duration) *ChannelTimer {
	return &ChannelTimer{
		C: make(chan time.Time, 1),
	}
}

func (t *ChannelTimer) Reset(d time.Duration) bool      {}
func After(d time.Duration) <-chan time.Time            {}
func AfterFunc(d time.Duration, f func()) *ChannelTimer {}
