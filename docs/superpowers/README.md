# Superpowers Specs And Plans

本目录保存经过确认的需求设计、实施计划和阶段性规格文档。

## 目录约定

- `specs/`：设计规格，记录目标、范围、非目标、信息架构、验收标准和风险。
- `plans/`：实施计划，记录可执行任务、涉及文件、验证命令和提交节奏。

## 当前文档

- `specs/2026-04-27-recharge-risk-design.md`：充值账号风控管理设计。
- `plans/2026-04-27-recharge-risk.md`：充值账号风控管理实施计划。
- `specs/2026-04-26-kakayun-maxmoney-loss-guard-design.md`：卡卡云下单 `maxmoney` 防亏本设计。
- `plans/2026-04-26-kakayun-maxmoney-loss-guard.md`：卡卡云下单 `maxmoney` 防亏本实施计划。

已完成事项应沉淀到稳定文档中，例如根 `README.md`、`docs/module-map.md`、`docs/development.md` 或 `docs/testing.md`。

## 维护要求

- spec 必须先被确认，再写 plan。
- plan 必须能被没有上下文的执行者按步骤推进。
- 已完成的实现结果应回写到根 `README.md` 或 `docs/*.md`，不要只停留在计划文档中。
