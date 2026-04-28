# Superpowers Specs And Plans

本目录保存经过确认的需求设计、实施计划和阶段性规格文档。

## 目录约定

- `specs/`：设计规格，记录目标、范围、非目标、信息架构、验收标准和风险。
- `plans/`：实施计划，记录可执行任务、涉及文件、验证命令和提交节奏。

## 当前文档

- `specs/2026-04-28-supplier-provider-full-integration-design.md`：全量供应商渠道下单、查单、商品同步、防亏损、拆单和订阅边界设计。
- `plans/2026-04-28-supplier-provider-full-integration.md`：全量供应商渠道对接实施计划，按 provider、订单 segment、商品同步和文档验证拆分任务。

已完成事项应沉淀到稳定文档中，例如根 `README.md`、`docs/module-map.md`、`docs/development.md` 或 `docs/testing.md`。

## 维护要求

- spec 必须先被确认，再写 plan。
- plan 必须能被没有上下文的执行者按步骤推进。
- 已完成的实现结果应回写到根 `README.md` 或 `docs/*.md`，不要只停留在计划文档中。
