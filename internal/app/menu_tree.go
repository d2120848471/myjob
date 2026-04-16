package app

import "sort"

// BuildMenuTree 将扁平菜单列表构建为树结构（按 sort/id 排序）。
func BuildMenuTree(items []AdminMenu, parentID int64) []*AdminMenu {
	grouped := make(map[int64][]*AdminMenu)
	for i := range items {
		item := &items[i]
		grouped[item.ParentID] = append(grouped[item.ParentID], item)
	}
	var walk func(int64) []*AdminMenu
	walk = func(pid int64) []*AdminMenu {
		current := grouped[pid]
		sort.Slice(current, func(i, j int) bool {
			if current[i].Sort == current[j].Sort {
				return current[i].ID < current[j].ID
			}
			return current[i].Sort < current[j].Sort
		})
		for _, node := range current {
			node.Children = walk(node.ID)
		}
		return current
	}
	return walk(parentID)
}
