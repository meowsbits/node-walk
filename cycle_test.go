package cycle

import (
	"errors"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
)

func newTestNode() *Node {
	return &Node{
		Name:     fake.FirstName(),
		Children: make([]*Node, 0),
		Cousins:  make(map[string][]Node),
	}
}

func newTestNodeName(name string) *Node {
	n := newTestNode()
	n.Name = name
	return n
}

func onNode(t *testing.T, walker *walker) func(node *Node) error {
	return func(node *Node) error {
		node.Met++
		treePre := strings.Repeat("\t", walker.depth)
		if walker.depth == 0 {
			treePre = ""
		}
		t.Log(spew.Sprintf("%d %s %+v", walker.depth, treePre, node))
		return nil
	}
}

func TestWalker_Walk(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		walker := newWalker()
		root := newTestNode().asRoot()

		err := walker.Walk(root, onNode(t, walker))
		assert.NoError(t, err)

		assert.Equal(t, 1, walker.iter)
		assert.Equal(t, 1, root.Met)
	})
	t.Run("child", func(t *testing.T) {
		walker := newWalker()

		root := newTestNode().asRoot()
		child := newTestNode()
		root.Children = append(root.Children, child)

		err := walker.Walk(root, onNode(t, walker))
		assert.NoError(t, err)

		assert.Equal(t, 2, walker.iter)
		assert.Equal(t, 1, root.Met)
		assert.Equal(t, 1, child.Met)
	})
	t.Run("returns error", func(t *testing.T) {
		walker := newWalker()

		root := newTestNode().asRoot()
		child := newTestNode()
		root.Children = append(root.Children, child)

		err := walker.Walk(root, func(node *Node) error {
			if walker.iter > 1 {
				return errors.New("myError")
			}
			return onNode(t, walker)(node)
		})
		assert.Error(t, err)

		assert.Equal(t, 2, walker.iter)
		assert.Equal(t, 0, root.Met)
		assert.Equal(t, 0, child.Met)
	})
	t.Run("children", func(t *testing.T) {
		walker := newWalker()

		root := newTestNode().asRoot()
		child := newTestNode()
		child2 := newTestNode()
		root.Children = append(root.Children, child)
		root.Children = append(root.Children, child2)

		err := walker.Walk(root, onNode(t, walker))
		assert.NoError(t, err)

		assert.Equal(t, 3, walker.iter)
		assert.Equal(t, 1, root.Met)
		assert.Equal(t, 1, child.Met)
		assert.Equal(t, 1, child2.Met)
	})
	t.Run("cousins", func(t *testing.T) {
		walker := newWalker()

		root := newTestNode().asRoot()
		child := newTestNode()
		child2 := newTestNode()
		root.Children = append(root.Children, child)
		root.Children = append(root.Children, child2) // iter:3
		for _, c := range root.Children {
			root.Cousins[c.Name] = append([]Node{}, *newTestNode(), *newTestNode(), *newTestNode()) // iter:3x3=9
			for i := range root.Cousins[c.Name] {

				// NOTE: We must assign to the element by index.
				root.Cousins[c.Name][i].Children = append([]*Node{}, newTestNode(), newTestNode()) //iter: 9 += 2[childen] * 3[cousins] (=6) * 2[cousins-children] = 12 => 9+12=21
			}
		}

		err := walker.Walk(root, onNode(t, walker))
		assert.NoError(t, err)
		assert.Equal(t, 21, walker.iter)
		assert.Equal(t, 1, root.Met)
		assert.Equal(t, 1, child.Met)
		assert.Equal(t, 1, child2.Met)
	})
	t.Run("cycle", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			walker := newWalker()

			root := newTestNode().asRoot()
			root.Children = append(root.Children, root)

			err := walker.Walk(root, onNode(t, walker))
			assert.NoError(t, err)
		})
		t.Run("multiple", func(t *testing.T) {
			walker := newWalker()

			root := newTestNode().asRoot()
			child := newTestNode()
			child2 := newTestNode()
			root.Children = append(root.Children, child)
			root.Children = append(root.Children, child2)
			for _, c := range root.Children {
				root.Cousins[c.Name] = append([]Node{}, *newTestNode(), *newTestNode(), *newTestNode())
				for i := range root.Cousins[c.Name] {
					root.Cousins[c.Name][i].Children = append([]*Node{}, newTestNode(), newTestNode(), root) // cycle
				}
			}

			err := walker.Walk(root, onNode(t, walker))
			assert.NoError(t, err)

			cycles := walker.cycles()

			assert.Len(t,  cycles, 6)
			assert.Equal(t, 21, walker.iter)
			assert.True(t, root.Cycling)
		})
		t.Run("nested", func(t *testing.T) {
			walker := newWalker()

			root := newTestNode().asRoot()

			abby := newTestNodeName("abby")
			bobby := newTestNodeName("bobby")
			charlie := newTestNodeName("charlie")

			root.Children = append(root.Children, abby)
			abby.Children = append(abby.Children, bobby)
			bobby.Children = append(bobby.Children, charlie)
			charlie.Children = append(charlie.Children, abby) // cycle

			err := walker.Walk(root, onNode(t, walker))
			assert.NoError(t, err)

			cycles := walker.cycles()

			assert.Len(t, cycles, 1)
			assert.True(t, abby.Cycling)

			assert.Equal(t, 1, root.Met)
			assert.Equal(t, 1, abby.Met)
			assert.Equal(t, 1, bobby.Met)
			assert.Equal(t, 1, charlie.Met)

			assert.Equal(t, 4, walker.iter)
		})
	})
}
