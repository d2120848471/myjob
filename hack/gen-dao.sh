#!/usr/bin/env bash
set -euo pipefail

# 统一从环境变量拿连接串，避免脚本里写死真实账号。
GF_DAO_LINK="${GF_DAO_LINK:-}"
if [ -z "$GF_DAO_LINK" ]; then
  echo "GF_DAO_LINK is required, e.g. mysql:root:root123456@tcp(127.0.0.1:3307)/admin?charset=utf8mb4&parseTime=true&loc=Local" >&2
  exit 1
fi

# 这里显式落到项目约定目录，避免生成结果重新散落回默认位置。
go run github.com/gogf/gf/cmd/gf/v2@v2.10.0 gen dao \
  -p . \
  -l "$GF_DAO_LINK" \
  -d internal/dao \
  -o internal/model/do \
  -e internal/model/entity \
  -tp internal/model/table \
  -s \
  -w \
  -v
