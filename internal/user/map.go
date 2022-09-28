package user

import (
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

type ordMap[Key comparable, Val, Data any] struct {
	data  map[Key]Data
	order []Key

	toKey  func(Data) Key
	toVal  func(Data) Val
	isLess func(Data, Data) bool
}

func newOrdMap[Key comparable, Val, Data any](
	key func(Data) Key,
	value func(Data) Val,
	less func(Data, Data) bool,
	data ...Data,
) ordMap[Key, Val, Data] {
	m := ordMap[Key, Val, Data]{
		data: make(map[Key]Data),

		toKey:  key,
		toVal:  value,
		isLess: less,
	}

	for _, d := range data {
		m.insert(d)
	}

	return m
}

func (set *ordMap[Key, Val, Data]) insert(data Data) {
	if _, ok := set.data[set.toKey(data)]; ok {
		set.delete(set.toKey(data))
	}

	set.data[set.toKey(data)] = data

	set.order = append(set.order, set.toKey(data))

	slices.SortFunc(set.order, func(a, b Key) bool {
		return set.isLess(set.data[a], set.data[b])
	})
}

func (set *ordMap[Key, Val, Data]) delete(key Key) Val {
	data, ok := set.data[key]
	if !ok {
		return *new(Val)
	}

	delete(set.data, key)

	set.order = xslices.Filter(set.order, func(otherKey Key) bool {
		return otherKey != key
	})

	return set.toVal(data)
}

func (set *ordMap[Key, Val, Data]) get(key Key) Val {
	return set.toVal(set.data[key])
}

func (set *ordMap[Key, Val, Data]) keys() []Key {
	return set.order
}

func (set *ordMap[Key, Val, Data]) values() []Val {
	return xslices.Map(set.order, func(key Key) Val {
		return set.toVal(set.data[key])
	})
}

func (set *ordMap[Key, Val, Data]) toMap() map[Key]Val {
	m := make(map[Key]Val)

	for _, key := range set.order {
		m[key] = set.toVal(set.data[key])
	}

	return m
}
