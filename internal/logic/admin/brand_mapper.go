package adminlogic

import "fmt"

func (l *BrandLogic) buildBrandCreateLog(level int, name string, parentID int64) string {
	levelLabel := brandLevelLabel(level)
	if parentID == 0 {
		return fmt.Sprintf("添加%s品牌：%s", levelLabel, name)
	}
	return fmt.Sprintf("添加%s品牌：%s（父级ID=%d）", levelLabel, name, parentID)
}

func brandLevelLabel(level int) string {
	switch level {
	case 1:
		return "一级"
	case 2:
		return "二级"
	case 3:
		return "三级"
	default:
		return fmt.Sprintf("%d级", level)
	}
}
