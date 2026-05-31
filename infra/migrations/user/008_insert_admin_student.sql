-- Migration: 008_insert_admin_student
-- Schema: users
-- Description: Garante a existência do registro de estudante do usuário administrador

INSERT INTO users.students (
    id,
    display_name,
    username,
    email,
    xp_total,
    streak_count,
    readiness_score,
    onboarding_completed,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-4000-a000-000000000000',
    'Administrador',
    'admin',
    'admin@preuni.com',
    0,
    0,
    100,
    true,
    now(),
    now()
)
ON CONFLICT (id) DO NOTHING;
