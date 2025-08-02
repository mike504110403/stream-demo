#!/bin/bash

# 測試腳本
echo "🚀 開始運行後端測試..."

# 設置顏色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 檢查是否在正確的目錄
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ 錯誤：請在 backend 目錄下運行此腳本${NC}"
    exit 1
fi

# 清理之前的測試文件
echo -e "${BLUE}🧹 清理之前的測試文件...${NC}"
rm -f coverage.out coverage.html

# 運行單元測試
echo -e "${BLUE}📋 運行單元測試...${NC}"
go test ./test/... -v -coverprofile=coverage.out

# 檢查測試結果
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 單元測試通過${NC}"
else
    echo -e "${RED}❌ 單元測試失敗${NC}"
    exit 1
fi

# 生成覆蓋率報告
echo -e "${BLUE}📊 生成覆蓋率報告...${NC}"
go tool cover -html=coverage.out -o coverage.html

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 覆蓋率報告生成成功${NC}"
    echo -e "${YELLOW}📄 覆蓋率報告：coverage.html${NC}"
else
    echo -e "${RED}❌ 覆蓋率報告生成失敗${NC}"
fi

# 顯示覆蓋率統計
echo -e "${BLUE}📈 覆蓋率統計：${NC}"
go tool cover -func=coverage.out

# 運行特定服務測試
echo -e "${BLUE}🎯 運行核心服務測試...${NC}"

# 測試直播服務
echo -e "${YELLOW}📺 測試直播服務...${NC}"
go test ./test/live_service_test.go -v

# 測試用戶服務
echo -e "${YELLOW}👤 測試用戶服務...${NC}"
go test ./test/user_service_test.go -v

# 測試工具函數
echo -e "${YELLOW}🔧 測試工具函數...${NC}"
go test ./test/unit_test.go -v

# 測試 WebSocket
echo -e "${YELLOW}🌐 測試 WebSocket...${NC}"
go test ./test/websocket_test.go -v

# 運行集成測試
echo -e "${BLUE}🔗 運行集成測試...${NC}"
go test ./test/integration_test.go -v

echo -e "${GREEN}🎉 所有測試完成！${NC}"
echo -e "${YELLOW}💡 提示：打開 coverage.html 查看詳細覆蓋率報告${NC}" 