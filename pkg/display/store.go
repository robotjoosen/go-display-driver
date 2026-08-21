package display

import (
	"github.com/puzpuzpuz/xsync/v4"
)

type DisplayState struct {
	ScreenType ScreenType
	Data       any
	ListIndex  int
	ListLength int
}

type AnimationState struct {
	ScreenType ScreenType
	Radius     int
	Direction  int
}

const (
	AnimationMinRadius = 2
	AnimationMaxRadius = 32
)

type Store[T any] struct {
	data xsync.Map[int, T]
}

func NewStore[T any]() *Store[T] {
	return &Store[T]{
		data: *xsync.NewMapOf[int, T](),
	}
}

func (s *Store[T]) Get(display int) (T, bool) {
	return s.data.Load(display)
}

func (s *Store[T]) Set(display int, value T) {
	s.data.Store(display, value)
}

type AnimationStore struct {
	*Store[AnimationState]
}

func NewAnimationStore() *AnimationStore {
	return &AnimationStore{
		Store: NewStore[AnimationState](),
	}
}

func (s *AnimationStore) Tick(display int, screenType ScreenType) int {
	anim, ok := s.Get(display)
	if !ok || anim.ScreenType != screenType {
		anim = AnimationState{
			ScreenType: screenType,
			Radius:     AnimationMinRadius,
			Direction:  +1,
		}
		s.Set(display, anim)

		return anim.Radius
	}

	anim.Radius += anim.Direction
	if anim.Radius >= AnimationMaxRadius || anim.Radius <= AnimationMinRadius {
		anim.Direction = -anim.Direction
	}

	s.Set(display, anim)

	return anim.Radius
}
