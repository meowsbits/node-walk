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

var errCycle = errors.New("cycle detected")
func isCycleError(err error) bool {
	return err != nil && err == errCycle
}
func errPositiveAndIsNotCycle(err error) bool {
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

	// Detect cycle.
	v := reflect.ValueOf(root)
	ptr := v.Pointer()
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

	for i, item := range root.Children {
		err := w.Walk(item, mutate)
		if errPositiveAndIsNotCycle(err) {
			return err
		}
		if isCycleError(err) {
			item.Cycling = true
		}
		root.Children[i] = item
	}

	for _, list := range root.Cousins {
		for i, item := range list {
			err := w.Walk(&item, mutate)
			if errPositiveAndIsNotCycle(err) {
				return err
			}
			if isCycleError(err) {
				item.Cycling = true
			}
			list[i] = item
		}
	}

	return mutate(root)
}
