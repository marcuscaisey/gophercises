package sitemap

type hashSet[T comparable] struct {
	elements map[T]struct{}
}

func newHashSet[T comparable]() *hashSet[T] {
	return &hashSet[T]{
		elements: map[T]struct{}{},
	}
}

func (s *hashSet[T]) Exists(element T) bool {
	_, exists := s.elements[element]
	return exists
}

func (s *hashSet[T]) Add(element T) {
	s.elements[element] = struct{}{}
}
