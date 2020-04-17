package cycle

import (
	"errors"
	"reflect"
)

type Node struct {
	IsRoot     bool
	Name       string
	Cycling    bool
	Met        int
	generation int // aka depth
	Children   []*Node
	Cousins    map[string][]Node
}

func (n *Node) asRoot() *Node {
	n.IsRoot = true
	return n
}

type walker struct {
	iter     int
	depth    int
	pointers map[uintptr]int
	cycle    []cycle
}

func newWalker() *walker {
	return &walker{
		depth:    -1,
		pointers: make(map[uintptr]int),
	}
}

type cycle struct {
	iter, depth int
}

func (w *walker) cycles() []cycle {
	return w.cycle
}

// errCycle is returned when a cycle is detected.
// A cycle is when fields (or fields of fields... of fields)
// are found to contain duplicate addresses.
var errCycle = errors.New("cycle detected")

// errIsCyle tells us if the err is equivalent to a cycle error.
func errIsCycle(err error) bool {
	return err == errCycle
}

// errYesCycleNo tells us if the error is not nil and not a cycle indicator.
func errYesCycleNo(err error) bool {
	return err != nil && err != errCycle
}

func (w *walker) Walk(root *Node, mutate func(node *Node) error) error {

	// Remove pointers at or below the current depth from map used to detect
	// circular refs.
	for k, depth := range w.pointers {
		if depth >= w.depth {
			delete(w.pointers, k)
		}
	}

	// Detect cycles.
	ptr := reflect.ValueOf(root).Pointer()
	if pDepth, ok := w.pointers[ptr]; ok && pDepth < w.depth {
		w.cycle = append(w.cycle, cycle{w.iter, w.depth})
		return errCycle
	}
	w.pointers[ptr] = w.depth

	w.iter++
	w.depth++
	defer func() {
		w.depth--
	}()

	for _, item := range root.Children {
		err := w.Walk(item, mutate)
		if errYesCycleNo(err) {
			return err
		}
		if errIsCycle(err) {
			item.Cycling = true
		}
	}

	for _, list := range root.Cousins {
		for i, item := range list {
			err := w.Walk(&item, mutate)
			if errYesCycleNo(err) {
				return err
			}
			if errIsCycle(err) {
				item.Cycling = true
			}
			list[i] = item
		}
	}

	return mutate(root)
}
