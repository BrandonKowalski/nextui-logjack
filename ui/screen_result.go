package ui

type ScreenResult[T any] struct {
	Value  T
	Action Action
}

func quit[T any](value T) ScreenResult[T] {
	return ScreenResult[T]{
		Value:  value,
		Action: ActionQuit,
	}
}

func withAction[T any](value T, action Action) ScreenResult[T] {
	return ScreenResult[T]{
		Value:  value,
		Action: action,
	}
}
