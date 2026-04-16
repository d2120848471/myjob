package adminlogic

import "github.com/gogf/gf/v2/database/gdb"

func (l *IndustryLogic) rebuildIndustrySort(tx gdb.TX, orderedIDs []int64) error {
	for index, id := range orderedIDs {
		if _, err := tx.Exec(`UPDATE product_industry SET sort = ?, updated_at = ? WHERE id = ?`, index+1, l.core.Now(), id); err != nil {
			return err
		}
	}
	return nil
}

func (l *IndustryLogic) rebuildIndustryBrandSort(tx gdb.TX, industryID int64, orderedBrandIDs []int64) error {
	for index, brandID := range orderedBrandIDs {
		if _, err := tx.Exec(`UPDATE product_industry_brand SET sort = ? WHERE industry_id = ? AND brand_id = ?`, index+1, industryID, brandID); err != nil {
			return err
		}
	}
	return nil
}
