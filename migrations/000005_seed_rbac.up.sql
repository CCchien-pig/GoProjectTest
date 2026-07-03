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

-- 根據剛插入的 roles 與 permissions，設定對應的 role_permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = '廠長' 
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = '工程師' AND p.name IN ('view:device', 'operate:device')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = '操作員' AND p.name IN ('view:device')
ON CONFLICT DO NOTHING;
