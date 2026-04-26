# Superpowers Specs And Plans

本目录保存经过确认的需求设计、实施计划和阶段性规格文档。

## 目录约定

- `specs/`：设计规格，记录目标、范围、非目标、信息架构、验收标准和风险。
- `plans/`：实施计划，记录可执行任务、涉及文件、验证命令和提交节奏。

## 当前文档

- `specs/2026-04-24-docs-redesign-design.md`：文档体系重构设计。
- `specs/2026-04-25-kakayun-product-sync-design.md`：卡卡云商品信息定时同步设计。
- `specs/2026-04-26-kakayun-product-push-design.md`：卡卡云商品价格推送、订阅记录和改价记录设计；2026-04-26 产品口径调整后，接收 URL 由运营在卡卡云后台配置，系统不再自动维护 `geturl/seturl`。
- `plans/2026-04-24-docs-redesign.md`：文档体系重构实施计划。
- `plans/2026-04-26-kakayun-product-push.md`：卡卡云商品价格推送、订阅记录和改价记录实施计划；其中接收 URL 自动维护步骤已被后续产品口径废弃，当前实现以 `docs/development.md` 和 `docs/module-map.md` 为准。

## 维护要求

- spec 必须先被确认，再写 plan。
- plan 必须能被没有上下文的执行者按步骤推进。
- 已完成的实现结果应回写到根 `README.md` 或 `docs/*.md`，不要只停留在计划文档中。
