package adminlogic

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/consts"
)

type productBrandTreeRow struct {
	ID       int64  `db:"id"`
	ParentID int64  `db:"parent_id"`
	Name     string `db:"name"`
	Sort     int    `db:"sort"`
}

// FormOptions 返回商品表单所需的选项数据（品牌树、模板、策略、主体等）。
func (l *ProductGoodsLogic) FormOptions(ctx context.Context, _ *adminapi.ProductGoodsFormOptionsReq) (*adminapi.ProductGoodsFormOptionsRes, error) {
	brandRows, err := l.loadBrandRows(ctx)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}

	templateRows := make([]struct {
		ID    int64  `db:"id"`
		Title string `db:"title"`
	}, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &templateRows, `SELECT id, title FROM product_template ORDER BY id DESC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}
	templates := make([]adminapi.ProductGoodsTemplateOption, 0, len(templateRows))
	for _, row := range templateRows {
		templates = append(templates, adminapi.ProductGoodsTemplateOption{ID: row.ID, Title: row.Title})
	}

	strategyRows := make([]struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &strategyRows, `SELECT id, name FROM product_purchase_limit_strategy WHERE status = ? ORDER BY id DESC`, consts.StatusEnabled); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}
	strategies := make([]adminapi.ProductGoodsStrategyOption, 0, len(strategyRows))
	for _, row := range strategyRows {
		strategies = append(strategies, adminapi.ProductGoodsStrategyOption{ID: row.ID, Name: row.Name})
	}

	subjectRows := make([]struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &subjectRows, `SELECT id, name FROM admin_subject WHERE has_tax = 1 ORDER BY id DESC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}
	subjects := make([]adminapi.ProductGoodsSubjectOption, 0, len(subjectRows))
	for _, row := range subjectRows {
		subjects = append(subjects, adminapi.ProductGoodsSubjectOption{ID: row.ID, Name: row.Name})
	}

	return &adminapi.ProductGoodsFormOptionsRes{
		Brands:                  buildProductBrandTree(brandRows),
		Templates:               templates,
		PurchaseLimitStrategies: strategies,
		Subjects:                subjects,
		GoodsTypes: []adminapi.ProductGoodsStringOption{
			{Value: productGoodsTypeCardSecret, Label: productGoodsTypeLabels[productGoodsTypeCardSecret]},
			{Value: productGoodsTypeDirectRecharge, Label: productGoodsTypeLabels[productGoodsTypeDirectRecharge]},
		},
		SupplyTypes: []adminapi.ProductGoodsStringOption{
			{Value: productGoodsSupplyTypeChannel, Label: "渠道供货"},
		},
		BooleanOptions: []adminapi.ProductGoodsIntOption{
			{Value: 1, Label: "是"},
			{Value: 0, Label: "否"},
		},
		StatusOptions: []adminapi.ProductGoodsIntOption{
			{Value: 1, Label: "启用"},
			{Value: 0, Label: "停用"},
		},
	}, nil
}

func (l *ProductGoodsLogic) loadBrandRows(ctx context.Context) ([]productBrandTreeRow, error) {
	rows := make([]productBrandTreeRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT id, parent_id, name, sort FROM product_brand ORDER BY sort ASC, id ASC`); err != nil {
		return nil, err
	}
	return rows, nil
}

func buildProductBrandTree(rows []productBrandTreeRow) []adminapi.ProductGoodsBrandTreeItem {
	childrenByParent := make(map[int64][]productBrandTreeRow, len(rows))
	for _, row := range rows {
		childrenByParent[row.ParentID] = append(childrenByParent[row.ParentID], row)
	}

	var build func(parentID int64) []adminapi.ProductGoodsBrandTreeItem
	build = func(parentID int64) []adminapi.ProductGoodsBrandTreeItem {
		children := childrenByParent[parentID]
		items := make([]adminapi.ProductGoodsBrandTreeItem, 0, len(children))
		for _, child := range children {
			grandChildren := build(child.ID)
			items = append(items, adminapi.ProductGoodsBrandTreeItem{
				ID:       child.ID,
				Name:     child.Name,
				IsLeaf:   len(grandChildren) == 0,
				Children: grandChildren,
			})
		}
		return items
	}

	return build(0)
}

func expandBrandIDs(rows []productBrandTreeRow, rootID int64) []int64 {
	childrenByParent := make(map[int64][]int64, len(rows))
	exists := false
	for _, row := range rows {
		if row.ID == rootID {
			exists = true
		}
		childrenByParent[row.ParentID] = append(childrenByParent[row.ParentID], row.ID)
	}
	if !exists {
		return nil
	}
	result := make([]int64, 0, 8)
	queue := []int64{rootID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		queue = append(queue, childrenByParent[current]...)
	}
	return result
}
