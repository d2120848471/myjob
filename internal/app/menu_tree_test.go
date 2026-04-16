package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildMenuTree_ReturnsDistinctNodes(t *testing.T) {
	items := []AdminMenu{
		{ID: 1, ParentID: 0, Sort: 1},
		{ID: 2, ParentID: 0, Sort: 2},
	}

	tree := BuildMenuTree(items, 0)
	require.Len(t, tree, 2)
	require.False(t, tree[0] == tree[1])
	require.Equal(t, int64(1), tree[0].ID)
	require.Equal(t, int64(2), tree[1].ID)
}

func TestBuildMenuTree_BuildsChildrenAndSorts(t *testing.T) {
	items := []AdminMenu{
		{ID: 1, ParentID: 0, Sort: 2},
		{ID: 2, ParentID: 0, Sort: 1},
		{ID: 3, ParentID: 2, Sort: 1},
	}

	tree := BuildMenuTree(items, 0)
	require.Len(t, tree, 2)
	require.Equal(t, int64(2), tree[0].ID)
	require.Equal(t, int64(1), tree[1].ID)
	require.Len(t, tree[0].Children, 1)
	require.Equal(t, int64(3), tree[0].Children[0].ID)
	require.Empty(t, tree[1].Children)
}
