package k

import (
	"fmt"
	"testing"
)

// TreeNode 是 TreeNodeInterface 的一个实现
type TreeNode struct {
	Id       int64               `json:"id"`       // 主键id
	ParentId int64               `json:"parentId"` // 父节点id
	Children []TreeNodeInterface `json:"children"` // 子节点列表
	Name     string              `json:"name"`     // 标题
	IsLeaf   bool                `json:"isLeaf"`   // 是否为叶子节点，true表示是叶子节点(没有子节点了)，false表示不是叶子节点(还有子节点)
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

func (n *TreeNode) SetIsLeaf(leaf bool) {
	n.IsLeaf = leaf
}

func TestTree(t *testing.T) {
	items := []*TreeNode{
		{Id: 1, ParentId: 0, Name: "你好1"},
		{Id: 2, ParentId: 1, Name: "你好2"},
		{Id: 3, ParentId: 1, Name: "你好3"},
		{Id: 4, ParentId: 2, Name: "你好4"},
	}

	roots := BuildTree(items, 0)
	/**
	[{"id":1,"parentId":0,"children":[
		{
			"id":2,
			"parentId":1,
			"children":[
				{"id":4,"parentId":2,"children":null,"name":"你好4","isLeaf":true}
			],
			"name":"你好2",
			"isLeaf":false
		},
		{"id":3,"parentId":1,"children":null,"name":"你好3","isLeaf":true}],"name":"你好1","isLeaf":false}]


	*/
	fmt.Println(MapToString(roots))
}
