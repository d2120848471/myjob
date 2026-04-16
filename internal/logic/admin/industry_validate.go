package adminlogic

import (
	"context"

	"myjob/internal/consts"
)

func (l *IndustryLogic) validateTopLevelBrands(ctx context.Context, brandIDs []int64) error {
	if len(brandIDs) == 0 {
		return nil
	}
	args := make([]any, 0, len(brandIDs))
	for _, brandID := range brandIDs {
		args = append(args, brandID)
	}
	count, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_brand
WHERE id IN (`+sqlPlaceholders(len(brandIDs))+`)
  AND parent_id = 0
`, args...)
	if err != nil {
		return apiErr(consts.CodeInternalError, "品牌校验失败")
	}
	if count.Int() != len(brandIDs) {
		return apiErr(consts.CodeBadRequest, "行业仅允许关联一级品牌")
	}
	return nil
}
