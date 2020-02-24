package fifo

import "testing"

func TestFifoArray(t *testing.T) {
	cap := 23
	fi := NewInt(cap)
	t.Run("test-normal", func(t *testing.T) {
		for i := 0; i < 93; i++ {
			var idx int
			if i >= cap-1 {
				idx = cap - 1
			} else {
				idx = i
			}

			expect := i + 1
			fi.Add(expect)

			v := fi.Get()[idx]
			if v != expect {
				t.Errorf("expected value %d at index %d, is %d", expect, idx, v)
				t.FailNow()
			}
		}
	})
}
