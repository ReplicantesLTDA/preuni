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
    '$2a$12$QQm/tzjaNEwR7TUY4Z6JV.8Vecup6LNGC743SrbjHFHakQVjFxYMW',
    true,
    now()
)
-- Como o índice único usa lower(email), atualizamos a senha caso já exista
ON CONFLICT (lower(email))
DO UPDATE SET password_hash = EXCLUDED.password_hash;
