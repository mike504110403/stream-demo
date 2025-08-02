# 前端建置測試指南

## 概述

本專案的前端測試策略已簡化為**建置測試**，確保程式碼品質和建置穩定性。

## 測試策略

### 為什麼選擇建置測試？

1. **效率優先**: 建置測試比單元測試更快，更適合 CI/CD 流程
2. **實際需求**: 確保程式碼可以成功編譯和建置
3. **品質保證**: 通過 TypeScript 型別檢查和 ESLint 程式碼檢查
4. **維護成本低**: 不需要維護複雜的測試案例

## 測試內容

### 1. 依賴安裝測試
```bash
npm ci
```
- 確保所有依賴正確安裝
- 驗證 package-lock.json 一致性

### 2. 程式碼品質檢查
```bash
npm run lint
```
- ESLint 檢查程式碼風格
- 發現潛在的程式碼問題
- 確保程式碼一致性

### 3. 型別檢查
```bash
npm run type-check
```
- TypeScript 型別檢查
- 發現型別錯誤
- 確保型別安全

### 4. 建置測試
```bash
npm run build
```
- Vue 專案建置
- 確保所有元件可以正確編譯
- 生成生產環境檔案

## CI 工作流程

### GitHub Actions 設定

```yaml
# 前端建置測試
frontend-build:
  runs-on: ubuntu-latest
  name: Frontend Build Test
  
  steps:
  - uses: actions/checkout@v4
  
  - name: Set up Node.js
    uses: actions/setup-node@v4
    with:
      node-version: '18'
      cache: 'npm'
      cache-dependency-path: frontend/package-lock.json
  
  - name: Install dependencies
    working-directory: ./frontend
    run: npm ci
  
  - name: Run linting
    working-directory: ./frontend
    run: npm run lint
  
  - name: Run type checking
    working-directory: ./frontend
    run: npm run type-check
  
  - name: Build frontend
    working-directory: ./frontend
    run: npm run build
```

## 本地測試

### 完整測試流程
```bash
cd frontend

# 1. 安裝依賴
npm ci

# 2. 程式碼檢查
npm run lint

# 3. 型別檢查
npm run type-check

# 4. 建置測試
npm run build
```

### 單項測試
```bash
# 只檢查程式碼風格
npm run lint

# 只檢查型別
npm run type-check

# 只建置
npm run build
```

## 故障排除

### 常見問題

#### 1. ESLint 錯誤
```bash
# 自動修復
npm run lint -- --fix

# 檢查特定檔案
npx eslint src/components/MyComponent.vue
```

#### 2. TypeScript 錯誤
```bash
# 詳細錯誤資訊
npm run type-check -- --verbose

# 檢查特定檔案
npx vue-tsc --noEmit src/components/MyComponent.vue
```

#### 3. 建置錯誤
```bash
# 詳細建置資訊
npm run build -- --debug

# 檢查依賴
npm ls
```

### 錯誤類型

#### 程式碼風格錯誤
- 縮排不一致
- 未使用的變數
- 缺少分號
- 行長度超限

#### 型別錯誤
- 型別不匹配
- 缺少必要屬性
- 函數參數錯誤
- 介面定義錯誤

#### 建置錯誤
- 模組找不到
- 語法錯誤
- 依賴衝突
- 資源載入失敗

## 最佳實踐

### 1. 開發流程
```bash
# 開發時定期檢查
npm run lint && npm run type-check

# 提交前完整測試
npm run lint && npm run type-check && npm run build
```

### 2. 程式碼品質
- 遵循 ESLint 規則
- 使用 TypeScript 型別
- 保持程式碼整潔
- 定期重構

### 3. 依賴管理
- 使用 `npm ci` 而非 `npm install`
- 定期更新依賴
- 檢查安全漏洞
- 保持 package-lock.json 同步

## 監控和改進

### 1. 建置時間監控
- 追蹤建置時間變化
- 優化大型依賴
- 使用快取加速

### 2. 錯誤趨勢分析
- 統計常見錯誤類型
- 改進開發工具
- 提供更好的錯誤訊息

### 3. 持續改進
- 定期檢視測試策略
- 根據專案需求調整
- 保持測試流程簡潔

## 總結

前端建置測試提供了一個平衡的解決方案：
- ✅ **快速執行**: 比單元測試更快
- ✅ **品質保證**: 確保程式碼可以正確建置
- ✅ **維護簡單**: 不需要維護測試案例
- ✅ **實際有效**: 符合專案實際需求

這種策略特別適合快速迭代的專案，既保證了程式碼品質，又不會拖慢開發速度。 