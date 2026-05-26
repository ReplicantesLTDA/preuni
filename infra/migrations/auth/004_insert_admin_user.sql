-- Migration: 004_insert_admin_user
-- Schema: auth
-- Description: Garante a existência de um usuário administrador inicial

-- Inserção do usuário administrador (E-mail: admin@preuni.com / Senha original: Admin123)
INSERT INTO auth.credentials (
    id,
    email,
    password_hash,
    email_verified,
    email_verified_at
)
VALUES (
    '00000000-0000-4000-a000-000000000000',
    'admin@preuni.com',
    '$2a$12$R9hZ7lBwO2B3qE7X9v8uGeUqgZf7W8yq8XoO8M1Xw6k1Gz5x6Y7uG',
    true,
    now()
)
-- Como o índice único usa lower(email), criamos uma estratégia para o conflito
ON CONFLICT (lower(email))
DO NOTHING;
