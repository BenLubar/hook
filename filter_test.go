package hook_test

import (
	"errors"
	"testing"

	"github.com/BenLubar/hook"
)

func expectPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("expected panic")
	} else {
		t.Logf("got panic: %v", r)
	}
}

func TestNewFilterPanic(t *testing.T) {
	t.Parallel()

	t.Run("NotPointer", func(t *testing.T) {
		t.Parallel()

		defer expectPanic(t)

		var x func() error

		_ = hook.NewFilter(x)
	})
	t.Run("NotFunc", func(t *testing.T) {
		t.Parallel()

		defer expectPanic(t)

		var x int

		_ = hook.NewFilter(&x)
	})
	t.Run("NoErrorResult", func(t *testing.T) {
		t.Parallel()

		defer expectPanic(t)

		var x func()

		_ = hook.NewFilter(&x)
	})
	t.Run("MismatchedTypes", func(t *testing.T) {
		t.Parallel()

		defer expectPanic(t)

		var x func(int) (string, error)

		_ = hook.NewFilter(&x)
	})
	t.Run("TooManyResults", func(t *testing.T) {
		t.Parallel()

		defer expectPanic(t)

		var x func(int) (int, int, error)

		_ = hook.NewFilter(&x)
	})
	t.Run("NoArgs", func(t *testing.T) {
		t.Parallel()

		var x func() error

		_ = hook.NewFilter(&x).(func(func() error, int))
	})
	t.Run("TypedNoArgs", func(t *testing.T) {
		t.Parallel()

		type X func() error

		var x X

		_ = hook.NewFilter(&x).(func(X, int))
	})
	t.Run("T2", func(t *testing.T) {
		t.Parallel()

		var x func(int, string) (int, string, error)

		_ = hook.NewFilter(&x).(func(func(int, string) (int, string, error), int))
	})
	t.Run("A3", func(t *testing.T) {
		t.Parallel()

		var x func(int, string, bool) error

		_ = hook.NewFilter(&x).(func(func(int, string, bool) error, int))
	})
	t.Run("T2A3", func(t *testing.T) {
		t.Parallel()

		var x func(int, string, int, string, bool) (int, string, error)

		_ = hook.NewFilter(&x).(func(func(int, string, int, string, bool) (int, string, error), int))
	})
}

func checkOrder(t *testing.T, order int) func(int) (int, error) {
	return func(i int) (int, error) {
		if order != i {
			t.Errorf("expected %d, got %d", order, i)
		}
		return i + 1, nil
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()

	t.Run("Order", func(t *testing.T) {
		t.Parallel()

		var apply func(int) (int, error)
		register := hook.NewFilter(&apply).(func(func(int) (int, error), int))
		register(checkOrder(t, 0), 0)
		register(checkOrder(t, 2), 1)
		register(checkOrder(t, 1), 0)

		n, err := apply(0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 3 {
			t.Errorf("expected result: 3, actual result: %d", n)
		}
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		var called bool

		defer func() {
			if !called {
				t.Errorf("first registered function was not called")
			}
		}()

		expectedErr := errors.New("[expected error]")

		var apply func(int) (int, error)
		register := hook.NewFilter(&apply).(func(func(int) (int, error), int))

		register(func(i int) (int, error) {
			called = true
			return i + 1, nil
		}, 0)
		register(func(i int) (int, error) {
			return i + 1, expectedErr
		}, 1)
		register(func(i int) (int, error) {
			t.Errorf("registered function after error should not be called")
			return i + 1, nil
		}, 2)

		n, err := apply(0)
		if err == nil {
			t.Errorf("expected error")
		} else if err != expectedErr {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 2 {
			t.Errorf("expected result: 2, actual result: %d", n)
		}
	})

	t.Run("Panic", func(t *testing.T) {
		t.Parallel()

		type panicTest struct{}

		var called bool

		defer func() {
			if !called {
				t.Errorf("first registered function was not called")
			}

			if r := recover(); r == nil {
				t.Errorf("expected panic")
			} else if r != (panicTest{}) {
				panic(r)
			}
		}()

		var apply func() error
		register := hook.NewFilter(&apply).(func(func() error, int))

		register(func() error {
			called = true
			return nil
		}, 0)
		register(func() error {
			panic(panicTest{})
		}, 1)
		register(func() error {
			t.Errorf("registered function after panic should not be called")
			return nil
		}, 2)

		err := apply()
		t.Errorf("unexpected return: %v", err)
	})
}
