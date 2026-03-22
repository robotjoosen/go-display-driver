package display

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robotjoosen/go-display-driver/pkg/state"
)

const (
	stateDebounceDelay = 5 * time.Second
	statePersistPeriod = 30 * time.Second
)

type Event interface {
	isEvent()
}

type RefreshEvent struct {
	Display int
}

type ControlEvent struct {
	KeyID     int
	Action    string
	Timestamp int64
}

type SelectEvent struct{}

type NextDisplayEvent struct{}

type PrevDisplayEvent struct{}

type ListUpEvent struct{}

type ListDownEvent struct{}

func (RefreshEvent) isEvent()     {}
func (ControlEvent) isEvent()     {}
func (SelectEvent) isEvent()      {}
func (NextDisplayEvent) isEvent() {}
func (PrevDisplayEvent) isEvent() {}
func (ListUpEvent) isEvent()      {}
func (ListDownEvent) isEvent()    {}

type Manager struct {
	mu              sync.RWMutex
	panel           Panel
	displays        map[int]DisplayState
	selectedIndex   int
	displayList     []int
	lastRender      map[int]time.Time
	eventQueue      chan Event
	keyState        map[int]string
	keyRepeatMs     int
	refreshInterval time.Duration
	stopChan        chan struct{}
	stateTimer      *time.Timer
	statePersist    time.Time
	stateDirty      bool
}

type DisplayState struct {
	ScreenType ScreenType
	Data       any
	ListIndex  int
	ListLength int
}

func NewManager(displays []int, p Panel) *Manager {
	displayMap := make(map[int]DisplayState)
	lastRenderMap := make(map[int]time.Time)
	for _, d := range displays {
		displayMap[d] = DisplayState{}
		lastRenderMap[d] = time.Time{}
	}

	m := &Manager{
		panel:           p,
		displays:        displayMap,
		displayList:     displays,
		lastRender:      lastRenderMap,
		eventQueue:      make(chan Event, 100),
		keyState:        make(map[int]string),
		keyRepeatMs:     200,
		refreshInterval: time.Duration(len(displays)) * 20 * time.Millisecond,
		stopChan:        make(chan struct{}),
	}

	go m.eventLoop()

	return m
}

func (m *Manager) Input(e Event) {
	select {
	case m.eventQueue <- e:
	default:
	}
}

func (m *Manager) eventLoop() {
	ticker := time.NewTicker(m.refreshInterval)
	defer ticker.Stop()
	stateTicker := time.NewTicker(statePersistPeriod)
	defer stateTicker.Stop()
	for {
		select {
		case e := <-m.eventQueue:
			m.handle(e)
		case <-ticker.C:
			m.refreshAll()
		case <-stateTicker.C:
			m.persistStateIfDirty()
		case <-m.stopChan:
			m.saveState()
			return
		}
	}
}

func (m *Manager) handle(e Event) {
	switch e := e.(type) {
	case RefreshEvent:
		m.queueRefresh(e.Display)

	case ControlEvent:
		m.handleControl(e)

	case SelectEvent:
		display := m.selectedDisplay()
		screen, ok := Get(m.getCurrentScreenType(display))
		if ok {
			if th, ok := screen.(TransitionHandler); ok {
				th.HandleSelect(display, m)
			}
		}

	case NextDisplayEvent:
		m.selectNext()
		m.queueRefresh(m.selectedDisplay())

	case PrevDisplayEvent:
		m.selectPrev()
		m.queueRefresh(m.selectedDisplay())

	case ListUpEvent:
		m.listUp(m.selectedDisplay())
		m.queueRefresh(m.selectedDisplay())

	case ListDownEvent:
		m.listDown(m.selectedDisplay())
		m.queueRefresh(m.selectedDisplay())
	}
}

func (m *Manager) Close() {
	close(m.stopChan)
}

func (m *Manager) handleControl(e ControlEvent) {
	display := m.selectedDisplay()
	keyID := e.KeyID

	slog.Debug("control event",
		"keyID", keyID,
		"action", e.Action,
		"KeyCycleDisplay", KeyCycleDisplay,
		"KeyCycleScreen", KeyCycleScreen,
		"KeyPrev", KeyPrev,
		"KeyNext", KeyNext,
		"KeySelect", KeySelect,
	)

	if e.Action == ActionRelease {
		delete(m.keyState, keyID)
		return
	} else {
		m.keyState[keyID] = e.Action
	}

	switch keyID {
	case KeyCycleDisplay:
		slog.Debug("key matched KeyCycleDisplay, calling selectNext")
		m.selectNext()
	case KeyCycleScreen:
		slog.Debug("key matched KeyCycleScreen, calling cycleScreenType")
		m.cycleScreenType(display)
	case KeyPrev:
		slog.Debug("key matched KeyPrev, calling listDown")
		m.listDown(display)
	case KeyNext:
		slog.Debug("key matched KeyNext, calling listUp")
		m.listUp(display)
	case KeySelect:
		slog.Debug("key matched KeySelect, calling HandleSelect")
		screen, ok := Get(m.getCurrentScreenType(display))
		if ok {
			if th, ok := screen.(TransitionHandler); ok {
				th.HandleSelect(display, m)
			}
		}
	default:
		slog.Warn("key not matched in switch", "keyID", keyID)
	}

	m.queueRefresh(m.selectedDisplay())
}

func (m *Manager) getCurrentScreenType(display int) ScreenType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.displays[display].ScreenType
}

func (m *Manager) selectedDisplay() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.displayList) == 0 {
		return 0
	}
	return m.displayList[m.selectedIndex]
}

func (m *Manager) DisplayList() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]int, len(m.displayList))
	copy(result, m.displayList)
	return result
}

func (m *Manager) selectNext() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.displayList) == 0 {
		return
	}
	m.selectedIndex = (m.selectedIndex + 1) % len(m.displayList)
	m.markStateDirty()
}

func (m *Manager) selectPrev() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.displayList) == 0 {
		return
	}
	m.selectedIndex = (m.selectedIndex - 1 + len(m.displayList)) % len(m.displayList)
	m.markStateDirty()
}

func (m *Manager) cycleScreenType(display int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.displays[display].ScreenType
	idx := -1
	for i, st := range ScreenTypeCycleOrder {
		if st == current {
			idx = i
			break
		}
	}

	nextIdx := (idx + 1) % len(ScreenTypeCycleOrder)
	state := m.displays[display]
	state.ScreenType = ScreenTypeCycleOrder[nextIdx]
	state.ListIndex = 0
	m.displays[display] = state

	slog.Debug("cycleScreenType",
		"display", display,
		"current", current,
		"idx", idx,
		"nextIdx", nextIdx,
		"nextScreen", ScreenTypeCycleOrder[nextIdx],
	)
	m.markStateDirty()
}

func (m *Manager) SetScreen(display int, screenType ScreenType, data any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.displays[display]
	state.ScreenType = screenType
	state.Data = data
	state.ListIndex = 0
	state.ListLength = 0
	m.displays[display] = state
	m.markStateDirty()
}

func (m *Manager) SetListLength(display int, length int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.displays[display]
	state.ListLength = length
	m.displays[display] = state
}

func (m *Manager) SetListIndex(display int, index int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.displays[display]
	if state.ListLength > 0 {
		state.ListIndex = index % state.ListLength
	}
	m.displays[display] = state
	m.markStateDirty()
}

func (m *Manager) listUp(display int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.displays[display]
	if state.ListLength <= 0 {
		return
	}
	state.ListIndex = (state.ListIndex - 1 + state.ListLength) % state.ListLength
	m.displays[display] = state
	m.markStateDirty()
}

func (m *Manager) listDown(display int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.displays[display]
	if state.ListLength <= 0 {
		return
	}
	state.ListIndex = (state.ListIndex + 1) % state.ListLength
	m.displays[display] = state
	m.markStateDirty()
}

func (m *Manager) GetState(display int) (DisplayState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.displays[display]
	return state, ok
}

func (m *Manager) refreshAll() {
	m.mu.RLock()
	displays := make([]int, len(m.displayList))
	copy(displays, m.displayList)
	m.mu.RUnlock()

	for _, d := range displays {
		m.queueRefresh(d)
	}
}

func (m *Manager) queueRefresh(display int) {
	m.mu.Lock()
	lastRender := m.lastRender[display]
	m.mu.Unlock()

	now := time.Now()
	if now.Sub(lastRender) < time.Duration(RefreshDebounceMs)*time.Millisecond {
		return
	}

	m.render(display)
}

func (m *Manager) render(display int) {
	m.mu.Lock()
	state, ok := m.displays[display]
	m.lastRender[display] = time.Now()
	m.mu.Unlock()

	if !ok {
		return
	}

	slog.Debug("render",
		"display", display,
		"screenType", state.ScreenType,
		"dataType", fmt.Sprintf("%T", state.Data),
	)

	screen, ok := Get(state.ScreenType)
	if !ok {
		return
	}

	m.panel.DisplayDraw(display, screen.Render(display, m))
}

func (m *Manager) LoadState() error {
	s, err := state.Load()
	if err != nil {
		return err
	}
	if s == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if s.SelectedIndex >= 0 && s.SelectedIndex < len(m.displayList) {
		m.selectedIndex = s.SelectedIndex
	}

	for displayID, snapshot := range s.Displays {
		if ds, ok := m.displays[displayID]; ok {
			ds.ScreenType = ScreenType(snapshot.ScreenType)
			ds.ListIndex = snapshot.ListIndex
			m.displays[displayID] = ds
		}
	}

	return nil
}

func (m *Manager) SaveState() error {
	m.mu.RLock()
	s := &state.State{
		SelectedIndex: m.selectedIndex,
		Displays:      make(map[int]state.DisplaySnapshot),
	}
	for displayID, ds := range m.displays {
		s.Displays[displayID] = state.DisplaySnapshot{
			ScreenType: string(ds.ScreenType),
			ListIndex:  ds.ListIndex,
		}
	}
	m.mu.RUnlock()

	return state.Save(s)
}

func (m *Manager) saveState() {
	if err := m.SaveState(); err != nil {
		slog.Error("failed to save state", "err", err)
	}
}

func (m *Manager) persistStateIfDirty() {
	m.mu.Lock()
	dirty := m.stateDirty
	m.mu.Unlock()

	if dirty {
		m.saveState()
		m.mu.Lock()
		m.stateDirty = false
		m.mu.Unlock()
	}
}

func (m *Manager) markStateDirty() {
	m.mu.Lock()
	m.stateDirty = true
	if m.stateTimer != nil {
		m.stateTimer.Stop()
	}
	m.stateTimer = time.AfterFunc(stateDebounceDelay, func() {
		m.persistStateIfDirty()
	})
	m.mu.Unlock()
}
