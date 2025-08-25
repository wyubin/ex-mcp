package maptool

import (
	"bytes"
	"encoding/json"
	"slices"
	"sort"

	"gopkg.in/yaml.v3"
)

type OrderedMap[T any] struct {
	Order []string
	Map   map[string]T
}

func (s *OrderedMap[T]) UnmarshalJSON(b []byte) error {
	json.Unmarshal(b, &s.Map)

	index := make(map[string]int)
	for key := range s.Map {
		s.Order = append(s.Order, key)
		esc, _ := json.Marshal(key) //Escape the key
		index[key] = bytes.Index(b, esc)
	}

	sort.Slice(s.Order, func(i, j int) bool { return index[s.Order[i]] < index[s.Order[j]] })
	return nil
}

func (s *OrderedMap[T]) UnmarshalYAML(value *yaml.Node) error {
	s.Order = []string{}
	s.Map = make(map[string]T)
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		var val T
		if err := value.Content[i].Decode(&key); err != nil {
			return err
		}
		if err := value.Content[i+1].Decode(&val); err != nil {
			return err
		}
		s.Order = append(s.Order, key)
		s.Map[key] = val
	}
	return nil
}

func (s OrderedMap[T]) MarshalJSON() ([]byte, error) {
	var b []byte
	count := 0
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')
	for _, key := range s.Order {
		item, found := s.Map[key]
		if !found {
			continue
		}
		km, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		// if not first item
		if count != 0 {
			buf.WriteRune(',')
		}
		buf.Write(km)
		buf.WriteRune(':')
		vm, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		buf.Write(vm)
		count += 1
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}

func (s OrderedMap[T]) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	for _, key := range s.Order {
		val, found := s.Map[key]
		if !found {
			continue
		}
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		}

		valNode := &yaml.Node{}
		if err := valNode.Encode(val); err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valNode)
	}

	return node, nil
}

func (s *OrderedMap[T]) Append(key string, val T) {
	if _, found := s.Map[key]; !found {
		s.Order = append(s.Order, key)
	}
	s.Map[key] = val
}

func (s *OrderedMap[T]) Pop(key string) (T, bool) {
	item, found := s.Map[key]
	if !found {
		return item, false
	}
	delete(s.Map, key)
	// remove from order
	res := []string{}
	for _, name := range s.Order {
		if name != key {
			res = append(res, name)
		}
	}
	s.Order = res
	return item, true
}

// orderBy
func (s *OrderedMap[T]) OrderBy(keys ...string) {
	if len(keys) == 0 {
		slices.Sort(s.Order)
		return
	}
	newOrder := []string{}
	for _, key := range keys {
		if _, found := s.Map[key]; found {
			newOrder = append(newOrder, key)
		}
	}
	s.Order = newOrder
}

func NewOrderedMap[T any]() OrderedMap[T] {
	return OrderedMap[T]{Order: []string{}, Map: make(map[string]T)}
}
