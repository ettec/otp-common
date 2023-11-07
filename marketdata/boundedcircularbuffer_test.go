package marketdata

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestCircularBufferAddAndRemove(t *testing.T) {

	b := newBoundedCircularBuffer[int32](4)

	in := []int32{1, 2, 3, 4}

	allAdded := true
	for _, val := range in {
		allAdded = b.addHead(val) && allAdded
	}

	if !allAdded {
		t.Fatal("expected all values to be added")
	}

	var out []int32

	i, ok := b.removeTail()
	for ok {
		out = append(out, i)
		i, ok = b.removeTail()
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatal("expected in to equal out")
	}

}

func TestGetTail(t *testing.T) {

	b := newBoundedCircularBuffer[int32](4)

	b.addHead(1)
	b.addHead(2)
	b.addHead(3)

	if val, ok := b.getTail(); !ok || val != 1 {
		t.FailNow()
	}

	b.removeTail()

	if val, ok := b.getTail(); !ok || val != 2 {
		t.FailNow()
	}

}

func TestCircularBufferDisallowsAddWhenFull(t *testing.T) {
	b := newBoundedCircularBuffer[int32](4)

	b.addHead(1)
	b.addHead(2)
	b.addHead(3)
	b.addHead(4)

	b.removeTail()
	b.removeTail()

	b.addHead(7)
	b.addHead(8)
	ok := b.addHead(9)
	if ok {
		t.Fatal("expected add to fail")
	}
}

func TestCircularBufferReturnsFalseWhenEmpty(t *testing.T) {

	b := newBoundedCircularBuffer[int32](4)

	b.addHead(1)
	b.addHead(2)
	b.removeTail()
	b.removeTail()
	b.addHead(3)
	b.addHead(4)
	b.addHead(6)
	b.addHead(7)

	b.removeTail()
	b.removeTail()
	b.removeTail()
	b.removeTail()
	_, ok := b.removeTail()

	if ok {
		t.Fatal("expected remove to fail")
	}

}

func TestCircularBufferReadOverCapacityBoundary(t *testing.T) {

	b := newBoundedCircularBuffer[int32](4)

	in := []int32{1, 2, 3, 4}

	for _, val := range in {
		b.addHead(val)
	}

	i, ok := b.removeTail()
	if i != 1 || !ok {
		t.FailNow()
	}

	i, ok = b.removeTail()
	if i != 2 || !ok {
		t.FailNow()
	}

	if !b.addHead(5) {
		t.FailNow()
	}

	if !b.addHead(6) {
		t.FailNow()
	}

	var out []int32

	i, ok = b.removeTail()
	for ok {
		out = append(out, i)
		i, ok = b.removeTail()
	}

	expected := []int32{3, 4, 5, 6}
	if !reflect.DeepEqual(expected, out) {
		t.Fatalf("expected out %v to equal %v", out, expected)
	}

}

func TestCircularBufferManyOperations(t *testing.T) {

	b := newBoundedCircularBuffer[int32](20)
	numOps := 10000
	var expectedOut []int32
	totalReads := 0
	for i := 0; i < numOps; i++ {

		if b.len < b.capacity {
			numAdds := rand.Intn(b.capacity - b.len)
			for j := 0; j < numAdds; j++ {
				r := rand.Int31n(100)
				ok := b.addHead(r)
				if !ok {
					t.Fatalf("expected add to be ok")
				}

				expectedOut = append(expectedOut, r)
			}
		}

		numReads := rand.Intn(b.len)
		for j := 0; j < numReads; j++ {
			_, ok := b.removeTail()
			if !ok {
				t.Fatalf("expected remove to be ok")
			}

			totalReads++
		}

	}

	var out []int32
	i, ok := b.removeTail()
	for ok {
		out = append(out, i)
		i, ok = b.removeTail()
	}

	expectedOut = expectedOut[totalReads:]
	if !reflect.DeepEqual(expectedOut, out) {
		t.Fatalf("expected out %v to equal %v", out, expectedOut)
	}

}
