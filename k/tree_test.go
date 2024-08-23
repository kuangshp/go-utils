package k

import (
	"fmt"
	"testing"
)

// TreeNode 是 TreeNodeInterface 的一个实现
type TreeNode struct {
	Id       int64               `json:"id"`
	ParentId int64               `json:"parentId"`
	Children []TreeNodeInterface `json:"children"`
	Name     string              `json:"name"`
}

// GetID 返回节点的ID（通过指针接收者）
func (n *TreeNode) GetID() int64 {
	return n.Id
}

// GetParentID 返回节点的ParentId（通过指针接收者）
func (n *TreeNode) GetParentID() int64 {
	return n.ParentId
}

// AddChild 向节点的子节点列表中添加一个子节点（通过指针接收者）
func (n *TreeNode) AddChild(child TreeNodeInterface) {
	n.Children = append(n.Children, child)
}

func TestTree(t *testing.T) {
	items := []*TreeNode{
		{Id: 1, ParentId: 0, Name: "你好1"},
		{Id: 2, ParentId: 1, Name: "你好2"},
		{Id: 3, ParentId: 1, Name: "你好3"},
		{Id: 4, ParentId: 2, Name: "你好4"},
	}

	roots := BuildTree(items)
	/**
	[{"id":1,"parentId":0,"children":[{"id":2,"parentId":1,"children":[{"id":4,"parentId":2,"children":null,"name":"你好4"}],"name":"你好2"},{"id":3,"pard":1,"children":null,"name":"你好3"}],"name":"你好1"}]
	*/
	fmt.Println(MapToString(roots))
}
