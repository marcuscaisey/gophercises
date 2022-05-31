package sitemap

type queueNode[T any] struct {
	value T
	next  *queueNode[T]
}

type linkedQueue[T any] struct {
	head *queueNode[T]
	tail *queueNode[T]
}

func (q *linkedQueue[T]) Enqueue(value T) {
	if q.head == nil && q.tail == nil {
		q.head = &queueNode[T]{value: value}
		q.tail = q.head
		return
	}

	newTail := &queueNode[T]{
		value: value,
	}
	q.tail.next = newTail
	q.tail = newTail
}

func (q *linkedQueue[T]) Dequeue() T {
	if q.Empty() {
		panic("dequeue of empty queue")
	}
	value := q.head.value
	if q.head == q.tail {
		q.head = nil
		q.tail = nil
	} else {
		nextHead := q.head.next
		q.head.next = nil
		q.head = nextHead
	}
	return value
}

func (q *linkedQueue[T]) Empty() bool {
	return q.head == nil && q.tail == nil
}
