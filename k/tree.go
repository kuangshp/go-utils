package k

// TreeNodeInterface 定义了树节点应该具有的方法
type TreeNodeInterface interface {
	GetID() int64
	GetParentID() int64
	AddChild(TreeNodeInterface)
}

// BuildTree 泛型化的构建树函数
func BuildTree[T TreeNodeInterface](nodes []T) []T {
	idMap := make(map[int64]T)
	var root []T

	// 初始化所有节点到 map 中
	for _, node := range nodes {
		idMap[node.GetID()] = node
	}

	// 构建树结构
	for _, node := range nodes {
		if node.GetParentID() == 0 {
			root = append(root, node) // 顶级节点
		} else if parent, exists := idMap[node.GetParentID()]; exists {
			parent.AddChild(node) // 添加为父节点的子节点
		}
	}

	return root
}
