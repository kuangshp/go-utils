package k

import "testing"

// TreeNode 是 TreeNodeInterface 的一个实现
type TreeNode struct {
	Id       int64               `json:"id"`
	ParentId int64               `json:"parentId"`
	Children []TreeNodeInterface `json:"children"`
	Name     string              `json:"name"`
	IsLeaf   bool                `json:"isLeaf"`
}

func (n *TreeNode) GetID() int64                   { return n.Id }
func (n *TreeNode) GetParentID() int64             { return n.ParentId }
func (n *TreeNode) AddChild(child TreeNodeInterface) { n.Children = append(n.Children, child) }
func (n *TreeNode) SetIsLeaf(leaf bool)            { n.IsLeaf = leaf }

func TestBuildTree(t *testing.T) {
	items := []*TreeNode{
		{Id: 1, ParentId: 0, Name: "根"},
		{Id: 2, ParentId: 1, Name: "子A"},
		{Id: 3, ParentId: 1, Name: "子B"},
		{Id: 4, ParentId: 2, Name: "孙"},
	}

	roots := BuildTree(items, 0)
	if len(roots) != 1 {
		t.Fatalf("根节点数 = %d", len(roots))
	}
	root := roots[0]
	if root.GetID() != 1 || root.IsLeaf {
		t.Fatalf("根节点异常: %+v", root)
	}
	if len(root.Children) != 2 {
		t.Fatalf("一级子节点 = %d", len(root.Children))
	}
	childA, ok := root.Children[0].(*TreeNode)
	if !ok || childA.Id != 2 || childA.IsLeaf || len(childA.Children) != 1 {
		t.Fatalf("子A异常: %+v", root.Children[0])
	}
}
