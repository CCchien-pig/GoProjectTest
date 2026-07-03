# Code Review：RBAC & User-Device 多對多架構變更

## 整體評分 ✅ 通過（附帶建議）

架構設計清晰、層次分明，主要的業務邏輯正確。以下依嚴重程度分類整理問題。

---

## 🔴 高優先級：需要修正

### 1. `users.role_id` 沒有 `NOT NULL` 約束
**檔案**: [`000001_create_users.up.sql`](file:///c:/Projects/CC/Go/GoProjectTest/migrations/000001_create_users.up.sql)

```diff
- role_id       UUID REFERENCES roles(id) ON DELETE SET NULL,
+ role_id       UUID NOT NULL REFERENCES roles(id),
```

**問題**：每個使用者都應該有一個角色，允許 `NULL` 代表無角色的使用者可以進入系統，這會在後來的鑑權邏輯中造成 `nil pointer` 或「無角色可判斷」的 edge case。如果確實有「無角色」的設計需求（如建立暫時帳號），才考慮保留 `NULL`。

> [!CAUTION]
> 若保留可 `NULL`，後續所有 Service/Handler 中都必須加 `user.Role == nil` 的防守判斷。

---

### 2. 建立/更新設備時無 Transaction 保護（N+1 查詢問題）
**檔案**: [`device_service.go`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go)

```go
// 目前做法：對每個 user_id 各自發一個 SELECT 查詢
for _, uid := range req.UserIDs {
    u, err := s.userRepo.FindByID(ctx, uid)  // N 次 DB 查詢
    ...
}
// 然後 repo.Create(device) 又獨立一個 TX
// 然後 repo.ReplaceUsers() 又獨立一個 TX
```

**問題 A（N+1）**：若 `UserIDs` 傳入 10 個 ID，會觸發 10 次獨立 SELECT。應改為批次查詢：
```go
// 建議：加一個 Repository 方法
FindByIDs(ctx, []uuid.UUID) ([]*model.User, error)
// 使用 WHERE id = ANY($1) 一次撈出
```

**問題 B（無事務）**：`Create(device)` 成功、但 `ReplaceUsers()` 失敗時，資料庫中會存在一個沒有使用者綁定的「孤兒設備」。應包在同一個 `db.Transaction` 中。

---

## 🟡 中優先級：建議改善

### 3. `Update` 中 `UserIDs` 的語意應文件化
**檔案**: [`device_service.go`](file:///c:/Projects/CC/Go/GoProjectTest/internal/service/device_service.go)

```go
if req.UserIDs != nil {  // nil 代表「不更新」，空陣列 [] 代表「清空所有使用者」
```

目前以 `req.UserIDs != nil` 作為「是否要更新關聯」的判斷，語意是正確的（遵循 PATCH 語意），但需要在 API 文件中明確記載：
- 傳 `null` 或省略此欄位 → 不異動設備綁定的使用者
- 傳 `[]` → 解除所有使用者綁定
- 傳 `["uuid1","uuid2"]` → **取代**（Replace，非新增）現有的所有綁定

> [!NOTE]
> 這個語意設計本身沒有問題，但需要 API 文件配合說明，否則前端開發者容易踩坑。

---

### 4. `roles` 和 `permissions` 表缺少 `updated_at`
**檔案**: [`000001_create_users.up.sql`](file:///c:/Projects/CC/Go/GoProjectTest/migrations/000001_create_users.up.sql)

若未來需要修改角色名稱或權限描述，會無法追蹤異動時間。建議統一加上以便稽核：

```sql
CREATE TABLE IF NOT EXISTS roles (
    ...
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()  -- 建議新增
);
```

---

### 5. `rbac.go` 中的 `RolePermission` struct 是多餘的
**檔案**: [`internal/model/rbac.go`](file:///c:/Projects/CC/Go/GoProjectTest/internal/model/rbac.go)

```go
// 這個 struct 目前沒有被任何地方使用
type RolePermission struct {
    RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
    PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}
```

GORM 的 `many2many:"role_permissions"` tag 已經自動管理這張中間表，定義此 struct 是多餘的。如果不需要對中間表執行獨立 CRUD，應將其刪除。

---

## 🟢 低優先級：優化建議

### 6. `device_repo.go` 的 `Preload` 重複
**檔案**: [`internal/repository/device_repo.go`](file:///c:/Projects/CC/Go/GoProjectTest/internal/repository/device_repo.go)

```go
// 目前寫法（重複 Preload Users 兩次）：
r.db.Preload("Users.Role.Permissions").Preload("Users").First(...)
```

GORM 中 `Preload("Users.Role.Permissions")` 已隱含 `Users` 的 Preload，不需要再加 `Preload("Users")`，可簡化為：
```go
r.db.Preload("Users.Role.Permissions").First(&device, "id = ?", id)
```

---

### 7. 缺乏 RBAC 初始種子資料 (Seed)

三張 RBAC 表建立後是空的，所有 User 的 `role_id` 初始時都沒有角色可選。建議新增一支 Seed Migration：

```sql
-- migrations/000005_seed_rbac.up.sql
INSERT INTO permissions (name, description) VALUES
    ('view:device',    '查看設備及遙測資料'),
    ('operate:device', '操作設備（修改狀態、推送命令）'),
    ('manage:system',  '管理系統（新增用戶、管理角色）')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (name, description) VALUES
    ('廠長',  '可完整管理系統與查看所有設備'),
    ('工程師', '可操作設備並查看遙測資料'),
    ('操作員', '只能查看設備狀態與遙測資料')
ON CONFLICT (name) DO NOTHING;
```

---

## 架構圖總結

```
✅ FK 設計正確：ON DELETE CASCADE / SET NULL 語意清晰
✅ user_devices 中間表結構乾淨，加了 created_at 可追蹤綁定時間
✅ Preload("Role.Permissions") 能完整帶出 RBAC 階層
✅ DTO 層次分離正確，RoleResp / PermissionResp 不洩漏 DB 內部結構
⚠️  主要風險：事務安全（Transaction）與 role_id NOT NULL 需要優先修正
```
