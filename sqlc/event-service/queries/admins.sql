-- Admin Management Queries

-- CREATE ADMIN
-- name: CreateAdmin :one
INSERT INTO admins (
    email, name, phone_number, password_hash, role, permissions
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING admin_id, email, name, phone_number, role, permissions, is_active, created_at;

-- GET ADMIN BY EMAIL (for login)
-- name: GetAdminByEmail :one
SELECT * FROM admins 
WHERE email = $1 AND is_active = true;

-- GET ADMIN BY ID
-- name: GetAdminByID :one
SELECT admin_id, email, name, phone_number, role, permissions, is_active, created_at, updated_at
FROM admins 
WHERE admin_id = $1;

-- UPDATE ADMIN PROFILE
-- name: UpdateAdminProfile :one
UPDATE admins 
SET name = COALESCE($2, name),
    phone_number = COALESCE($3, phone_number),
    updated_at = CURRENT_TIMESTAMP
WHERE admin_id = $1 AND is_active = true
RETURNING admin_id, email, name, phone_number, role, permissions, is_active, created_at, updated_at;

-- UPDATE ADMIN PERMISSIONS (Super Admin only)
-- name: UpdateAdminPermissions :one
UPDATE admins 
SET role = COALESCE($2, role),
    permissions = COALESCE($3, permissions),
    is_active = COALESCE($4, is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE admin_id = $1
RETURNING admin_id, email, name, phone_number, role, permissions, is_active, created_at, updated_at;

-- LIST ADMINS (Super Admin only)
-- name: ListAdmins :many
SELECT admin_id, email, name, phone_number, role, is_active, created_at
FROM admins
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- COUNT ADMINS
-- name: CountAdmins :one
SELECT COUNT(*) FROM admins;

-- DEACTIVATE ADMIN
-- name: DeactivateAdmin :exec
UPDATE admins 
SET is_active = false, updated_at = CURRENT_TIMESTAMP
WHERE admin_id = $1;

-- CHECK ADMIN PERMISSIONS
-- name: CheckAdminPermissions :one
SELECT admin_id, role, permissions, is_active
FROM admins 
WHERE admin_id = $1 AND is_active = true;