package display

import (
	"image"
	"image/color"
	"testing"
	"time"
)

type mockPanel struct {
	draws []drawCall
}

type drawCall struct {
	channel int
	img     image.Image
}

func (m *mockPanel) DisplayDraw(channel int, img image.Image) error {
	m.draws = append(m.draws, drawCall{channel, img})
	return nil
}

type simpleScreen struct{}

func (s *simpleScreen) Render(display int, m *Manager) image.Image {
	return image.NewGray(image.Rect(0, 0, 128, 64))
}

func TestNewManager(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0, 1}, p, "")

	if len(m.DisplayList()) != 2 {
		t.Errorf("expected 2 displays, got %d", len(m.DisplayList()))
	}

	state, ok := m.GetState(0)
	if !ok {
		t.Error("expected display 0 to exist")
	}
	if state.ScreenType != "" {
		t.Errorf("expected empty screen type, got %s", state.ScreenType)
	}
}

func TestSelectNextDisplay(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0, 1}, p, "")

	m.Input(NextDisplayEvent{})

	time.Sleep(10 * time.Millisecond)

	state, _ := m.GetState(0)
	if state.ScreenType != "" {
		t.Errorf("unexpected state change on non-selected display")
	}
}

func TestSelectNextWrap(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0, 1}, p, "")

	m.Input(NextDisplayEvent{})
	m.Input(NextDisplayEvent{})
	m.Input(NextDisplayEvent{})

	time.Sleep(10 * time.Millisecond)
}

func TestCycleScreenType(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0, 1}, p, "")

	m.SetScreen(0, ScreenDeviceStatus, nil)
	m.Input(ControlEvent{KeyID: KeyCycleScreen, Action: "PRESS"})

	time.Sleep(10 * time.Millisecond)

	state, _ := m.GetState(0)
	if state.ScreenType != ScreenNetworkStatus {
		t.Errorf("expected ScreenNetworkStatus, got %s", state.ScreenType)
	}
}

func TestListNavigation(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")

	m.SetScreen(0, ScreenDeviceStatus, nil)
	m.SetListLength(0, 5)
	m.SetListIndex(0, 2)

	m.Input(ControlEvent{KeyID: KeyNext, Action: "PRESS"})
	time.Sleep(10 * time.Millisecond)

	state, _ := m.GetState(0)
	if state.ListIndex != 1 {
		t.Errorf("expected ListIndex 1, got %d", state.ListIndex)
	}

	m.Input(ControlEvent{KeyID: KeyPrev, Action: "PRESS"})
	time.Sleep(10 * time.Millisecond)

	state, _ = m.GetState(0)
	if state.ListIndex != 2 {
		t.Errorf("expected ListIndex 2, got %d", state.ListIndex)
	}
}

func TestRefreshEvent(t *testing.T) {
	Register(ScreenDeviceStatus, &simpleScreen{})

	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)

	m.Input(RefreshEvent{Display: 0})
	time.Sleep(150 * time.Millisecond)

	if len(p.draws) == 0 {
		t.Error("expected at least one draw call")
	}
}

func TestClose(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")

	m.Close()

	m.Input(NextDisplayEvent{})
	time.Sleep(10 * time.Millisecond)
}

func TestControlEventKeyDown(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0, 1}, p, "")

	m.Input(ControlEvent{KeyID: KeyCycleDisplay, Action: "PRESS"})
	time.Sleep(10 * time.Millisecond)
}

type transitionHandlerScreen struct {
	called bool
}

func (s *transitionHandlerScreen) Render(display int, m *Manager) image.Image {
	return image.NewGray(image.Rect(0, 0, 128, 64))
}

func (s *transitionHandlerScreen) HandleSelect(display int, m *Manager) {
	s.called = true
}

func TestSelectKey(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)

	m.Input(ControlEvent{KeyID: KeySelect, Action: "PRESS"})
	time.Sleep(10 * time.Millisecond)
}

func TestControlEventRelease(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")

	m.Input(ControlEvent{KeyID: KeyCycleDisplay, Action: ActionRelease})
	time.Sleep(10 * time.Millisecond)
}

func TestSelectEvent(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)

	m.Input(SelectEvent{})
	time.Sleep(10 * time.Millisecond)
}

func TestListUpEvent(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)
	m.SetListLength(0, 5)
	m.SetListIndex(0, 2)

	m.Input(ListUpEvent{})
	time.Sleep(10 * time.Millisecond)

	state, _ := m.GetState(0)
	if state.ListIndex != 1 {
		t.Errorf("expected ListIndex 1, got %d", state.ListIndex)
	}
}

func TestListDownEvent(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)
	m.SetListLength(0, 5)
	m.SetListIndex(0, 2)

	m.Input(ListDownEvent{})
	time.Sleep(10 * time.Millisecond)

	state, _ := m.GetState(0)
	if state.ListIndex != 3 {
		t.Errorf("expected ListIndex 3, got %d", state.ListIndex)
	}
}

func TestPrevDisplayEvent(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0, 1}, p, "")

	m.Input(NextDisplayEvent{})
	m.Input(PrevDisplayEvent{})
	time.Sleep(10 * time.Millisecond)
}

func TestQueueRefreshDebounce(t *testing.T) {
	Register(ScreenDeviceStatus, &simpleScreen{})

	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)

	m.Input(RefreshEvent{Display: 0})
	m.Input(RefreshEvent{Display: 0})
	m.Input(RefreshEvent{Display: 0})
	time.Sleep(150 * time.Millisecond)

	if len(p.draws) == 0 {
		t.Error("expected at least one draw call")
	}
}

func TestSetListIndexBounds(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)
	m.SetListLength(0, 3)

	m.SetListIndex(0, 10)
	state, _ := m.GetState(0)
	if state.ListIndex != 1 {
		t.Errorf("expected ListIndex 1 (10%%3), got %d", state.ListIndex)
	}
}

func TestGetState(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")

	_, ok := m.GetState(999)
	if ok {
		t.Error("expected display 999 to not exist")
	}
}

func TestDisplayDrawError(t *testing.T) {
	p := &mockPanel{}
	m := NewManager([]int{0}, p, "")
	m.SetScreen(0, ScreenDeviceStatus, nil)

	img := image.NewGray(image.Rect(0, 0, 128, 64))
	img.Set(0, 0, color.White)

	m.Input(RefreshEvent{Display: 0})
	time.Sleep(10 * time.Millisecond)
}
