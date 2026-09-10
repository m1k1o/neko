package mediaqueue

import "testing"

func TestQueueEvictsOldestItemWhenFull(t *testing.T) {
	q := New[int](2, nil)

	q.Push(1)
	q.Push(2)
	q.Push(3)

	first, ok := q.Pop()
	if !ok || first != 2 {
		t.Fatalf("first item = (%d, %t), want (2, true)", first, ok)
	}
	second, ok := q.Pop()
	if !ok || second != 3 {
		t.Fatalf("second item = (%d, %t), want (3, true)", second, ok)
	}

	stats := q.Stats()
	if stats.Depth != 0 || stats.Dropped != 1 || stats.Capacity != 2 {
		t.Fatalf("stats = %+v, want capacity=2 depth=0 dropped=1", stats)
	}
}

func TestQueueCloseIsSafeWithPendingItems(t *testing.T) {
	q := New[int](1, nil)
	q.Push(1)
	q.Close()
	q.Close()

	if q.Push(2) {
		t.Fatal("Push succeeded after Close")
	}
	if _, ok := q.Pop(); ok {
		t.Fatal("Pop succeeded after Close")
	}
}
