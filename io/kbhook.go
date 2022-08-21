package io

import (
	"sync"

	"github.com/go-vgo/robotgo"
	gohook "github.com/robotn/gohook"
)

type KBInputHandler interface {
	InitKBHandler()
	GetEvents()
}

type KbInputHandler struct {
	EventMutex sync.Mutex
	Events     []gohook.Event
	Hooked     bool
}

func (ih *KbInputHandler) PushEvent(e gohook.Event) {
	ih.EventMutex.Lock()
	ih.Events = append(ih.Events, e)
	ih.EventMutex.Unlock()
}

func (ih *KbInputHandler) PopEvents() []gohook.Event {
	ih.EventMutex.Lock()
	tmp := ih.Events
	ih.Events = []gohook.Event{}
	ih.EventMutex.Unlock()

	return tmp
}

func (ih *KbInputHandler) GetEvents() []gohook.Event {
	if !ih.Hooked {
		ih.InitKBHandler()
	}

	return ih.PopEvents()
}

func (ih *KbInputHandler) InitKBHandler() {
	eventHook := robotgo.EventStart()
	var e gohook.Event
	ih.Hooked = true

	go func() {
		for e = range eventHook {
			if e.Kind == gohook.KeyDown {
				ih.PushEvent(e)
			}
		}
	}()
}
