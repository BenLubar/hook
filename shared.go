package hook

import (
	"reflect"
	"sort"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var intType = reflect.TypeOf(int(0))

type priorityValue struct {
	value    reflect.Value
	priority int
}

type priorityValues []priorityValue

func (pv priorityValues) Add(v reflect.Value, priority int) priorityValues {
	i := sort.Search(len(pv), func(i int) bool {
		return pv[i].priority > priority
	})

	added := make(priorityValues, len(pv)+1)
	copy(added, pv[:i])
	added[i] = priorityValue{v, priority}
	copy(added[i+1:], pv[i:])

	return added
}
