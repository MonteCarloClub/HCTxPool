package filter

// TODO make use of the generic filtering system

import (
	"sync"

	event2 "github.com/MonteCarloClub/KBD/model/event"
	"github.com/MonteCarloClub/KBD/model/state"

	"github.com/MonteCarloClub/KBD/filter"
)

type FilterManager struct {
	eventMux *event2.TypeMux

	filterMu sync.RWMutex
	filterId int
	filters  map[int]*filter.Filter

	quit chan struct{}
}

func NewFilterManager(mux *event2.TypeMux) *FilterManager {
	return &FilterManager{
		eventMux: mux,
		filters:  make(map[int]*filter.Filter),
	}
}

func (self *FilterManager) Start() {
	go self.filterLoop()
}

func (self *FilterManager) Stop() {
	close(self.quit)
}

func (self *FilterManager) InstallFilter(filter *filter.Filter) (id int) {
	self.filterMu.Lock()
	defer self.filterMu.Unlock()
	id = self.filterId
	self.filters[id] = filter
	self.filterId++

	return id
}

func (self *FilterManager) UninstallFilter(id int) {
	self.filterMu.Lock()
	defer self.filterMu.Unlock()
	if _, ok := self.filters[id]; ok {
		delete(self.filters, id)
	}
}

// GetFilter retrieves a filter installed using InstallFilter.
// The filter may not be modified.
func (self *FilterManager) GetFilter(id int) *filter.Filter {
	self.filterMu.RLock()
	defer self.filterMu.RUnlock()
	return self.filters[id]
}

func (self *FilterManager) filterLoop() {
	// Subscribe to events
	events := self.eventMux.Subscribe(
		//core.PendingBlockEvent{},
		event2.ChainEvent{},
		event2.TxPreEvent{},
		state.Logs(nil))

out:
	for {
		select {
		case <-self.quit:
			break out
		case e := <-events.Chan():
			switch e := e.(type) {
			case event2.ChainEvent:
				self.filterMu.RLock()
				for _, filter := range self.filters {
					if filter.BlockCallback != nil {
						filter.BlockCallback(e.Block, e.Logs)
					}
				}
				self.filterMu.RUnlock()

			case event2.TxPreEvent:
				self.filterMu.RLock()
				for _, filter := range self.filters {
					if filter.TransactionCallback != nil {
						filter.TransactionCallback(e.Tx)
					}
				}
				self.filterMu.RUnlock()

			case state.Logs:
				self.filterMu.RLock()
				for _, filter := range self.filters {
					if filter.LogsCallback != nil {
						msgs := filter.FilterLogs(e)
						if len(msgs) > 0 {
							filter.LogsCallback(msgs)
						}
					}
				}
				self.filterMu.RUnlock()
			}
		}
	}
}
