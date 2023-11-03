package util

import "sort"

type Set struct {
	data map[string]struct{}
}

func NewSet() *Set {
	return &Set{
		data: make(map[string]struct{}),
	}
}

func (s Set) Length() int {
	return len(s.data)
}

func (s *Set) Add(item string) {
	s.data[item] = struct{}{}
}

func (s Set) List() []string {
	out := []string{}
	for v := range s.data {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func (s *Set) Contains(item string) bool {
	_, found := s.data[item]
	return found
}

func (s *Set) Remove(item string) {
	delete(s.data, item)
}

func (s *Set) Intersection(other *Set) *Set {
	result := NewSet()
	for item := range s.data {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}
