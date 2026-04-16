package adminlogic

import "myjob/internal/app"

// BrandLogic 提供品牌管理相关业务能力（增删改查、排序、显隐、图片上传等）。
type BrandLogic struct{ core *app.Core }

// brand.go 仅保留 BrandLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - brand_query.go：列表/详情查询与基础数据加载
// - brand_write.go：新增/编辑/删除/排序/显隐写入逻辑
// - brand_upload.go：图片上传与本地存储落盘
// - brand_validate.go：父级层级、上传参数与文件校验
// - brand_mapper.go：日志描述与层级文案构建
