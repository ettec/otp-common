package marketdata

type boundedCircularBuffer[T any] struct {
	buffer   []T
	capacity int
	len      int
	readPtr  int
	writePtr int
}

func newBoundedCircularBuffer[T any](capacity int) *boundedCircularBuffer[T] {
	b := &boundedCircularBuffer[T]{buffer: make([]T, capacity), capacity: capacity}

	return b
}

// true if the buffer is not full and the value is added
func (b *boundedCircularBuffer[T]) addHead(item T) bool {

	if b.len == b.capacity {
		return false
	}

	b.buffer[b.writePtr] = item
	b.len++

	if b.writePtr == b.capacity-1 {
		b.writePtr = 0
	} else {
		b.writePtr++
	}

	return true

}

func (b *boundedCircularBuffer[T]) getTail() (T, bool) {
	var result T
	if b.len == 0 {
		return result, false
	}

	return b.buffer[b.readPtr], true
}

// returns the value and true if a value is available
func (b *boundedCircularBuffer[T]) removeTail() (T, bool) {
	var result T
	if b.len == 0 {
		return result, false
	}

	result = b.buffer[b.readPtr]
	b.len--
	b.readPtr++
	if b.readPtr == b.capacity {
		b.readPtr = 0
	}

	return result, true

}
