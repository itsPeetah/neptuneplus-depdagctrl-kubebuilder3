package signals

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

type StopSignalTable struct {
	signals sync.Map // map[types.NamespacedName]chan struct{}
}

func NewStopSignalTable() *StopSignalTable {
	return &StopSignalTable{
		signals: sync.Map{},
	}
}

func (t *StopSignalTable) Get(name types.NamespacedName) (chan struct{}, bool) {
	val, ok := t.signals.Load(name)
	if !ok {
		return nil, false
	}
	return val.(chan struct{}), true
}

func (t *StopSignalTable) Set(name types.NamespacedName, stopCh chan struct{}) {
	t.signals.Store(name, stopCh)
}

func (t *StopSignalTable) Delete(name types.NamespacedName) {
	if stopCh, ok := t.Get(name); ok {
		close(stopCh)
	}
	t.signals.Delete(name)
}

// Stops all recurring functions and clears the table.
func (t *StopSignalTable) Clear() {
	t.signals.Range(func(key, value interface{}) bool {
		name := key.(types.NamespacedName)
		t.Delete(name)
		return true // continue iteration
	})
}
