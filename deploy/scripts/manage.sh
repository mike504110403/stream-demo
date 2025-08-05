#!/bin/bash

# 簡化的專案管理腳本
# 作為 deploy/scripts/docker-manage.sh 的快捷方式

# 獲取腳本所在目錄
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 執行 docker 管理腳本
exec "$SCRIPT_DIR/docker-manage.sh" "$@" 