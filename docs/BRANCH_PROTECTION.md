# 分支保護策略設定指南

## 概述
為了管理公開專案的程式碼品質和安全性，需要設定分支保護規則。

## GitHub 分支保護設定

### 1. 進入分支保護設定
1. 前往 GitHub 專案頁面
2. 點擊 `Settings` 標籤
3. 左側選單選擇 `Branches`
4. 點擊 `Add rule` 或編輯現有規則

### 2. 主要分支保護規則

#### 針對 `main` 分支：
```
Branch name pattern: main
```

**保護設定：**
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: 2
  - ✅ Dismiss stale PR approvals when new commits are pushed
  - ✅ Require review from code owners
- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - ✅ Status checks: `backend-test`, `frontend-build`, `code-quality`
- ✅ **Require conversation resolution before merging**
- ✅ **Require signed commits**
- ✅ **Require linear history**
- ✅ **Include administrators**
- ✅ **Restrict pushes that create files that are larger than 100 MB**

#### 針對 `develop` 分支：
```
Branch name pattern: develop
```

**保護設定：**
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: 1
  - ✅ Dismiss stale PR approvals when new commits are pushed
- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - ✅ Status checks: `backend-test`, `frontend-build`, `code-quality`
- ✅ **Require conversation resolution before merging**
- ✅ **Include administrators**

### 3. 分支命名規範

#### 允許的分支類型：
- `feature/*` - 新功能開發
- `bugfix/*` - 錯誤修復
- `hotfix/*` - 緊急修復
- `release/*` - 版本發布
- `main` - 主分支
- `develop` - 開發分支

#### 分支命名範例：
```
feature/user-authentication
bugfix/login-validation
hotfix/security-patch
release/v1.2.0
```

### 4. 程式碼審查要求

#### 審查者設定：
- **main 分支**: 需要 2 個審查者批准
- **develop 分支**: 需要 1 個審查者批准
- **CODEOWNERS**: 自動指派相關檔案的所有者

#### CODEOWNERS 檔案設定：
創建 `.github/CODEOWNERS` 檔案：
```
# 後端相關
/backend/ @backend-team
/backend/api/ @api-team
/backend/services/ @services-team

# 前端相關
/frontend/ @frontend-team
/frontend/src/ @frontend-team

# 配置文件
*.yml @devops-team
*.yaml @devops-team
docker-compose*.yml @devops-team

# 文檔
/docs/ @documentation-team
README.md @maintainers
```

### 5. CI/CD 狀態檢查

#### 必須通過的檢查：
1. **backend-test** - 後端測試
   - Go 測試執行
   - 測試覆蓋率檢查 (≥30%)
   - 程式碼格式檢查
   - 建置測試

2. **frontend-build** - 前端建置
   - 依賴安裝
   - Linting 檢查
   - TypeScript 類型檢查
   - 建置測試

3. **code-quality** - 程式碼品質
   - 敏感檔案檢查
   - 檔案權限檢查
   - YAML 語法驗證

### 6. 提交訊息規範

#### 提交訊息格式：
```
<type>(<scope>): <subject>

<body>

<footer>
```

#### 類型 (type)：
- `feat`: 新功能
- `fix`: 錯誤修復
- `docs`: 文檔更新
- `style`: 程式碼格式調整
- `refactor`: 重構
- `test`: 測試相關
- `chore`: 建置過程或輔助工具的變動

#### 範例：
```
feat(auth): add JWT token validation

- Implement JWT token validation middleware
- Add token refresh functionality
- Update authentication tests

Closes #123
```

### 7. 安全設定

#### 敏感資訊保護：
- 禁止提交生產環境配置文件
- 禁止硬編碼密碼
- 使用環境變數管理敏感資訊

#### 檔案大小限制：
- 單一檔案不得超過 100MB
- 圖片檔案建議使用 Git LFS

### 8. 自動化工具

#### 建議的 GitHub Apps：
- **Dependabot**: 自動更新依賴
- **CodeQL**: 程式碼安全分析
- **Stale**: 自動關閉過期 PR/Issue

#### 設定範例：
```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
  - package-ecosystem: "gomod"
    directory: "/backend"
    schedule:
      interval: "weekly"
```

### 9. 緊急情況處理

#### 繞過保護規則：
- 只有管理員可以繞過分支保護
- 需要記錄繞過原因
- 事後需要補上審查

#### 緊急修復流程：
1. 創建 `hotfix/*` 分支
2. 修復問題
3. 創建 PR 到 `main`
4. 快速審查和合併
5. 合併到 `develop`

### 10. 監控和報告

#### 定期檢查項目：
- PR 審查時間
- CI 失敗率
- 程式碼覆蓋率趨勢
- 安全漏洞數量

#### 報告工具：
- GitHub Insights
- 自定義 GitHub Actions 報告
- 第三方整合工具

## 注意事項

1. **權限管理**: 定期檢查團隊成員權限
2. **規則更新**: 根據專案發展調整保護規則
3. **培訓**: 確保團隊了解分支保護規則
4. **備份**: 定期備份重要分支
5. **監控**: 監控 CI/CD 效能和穩定性 