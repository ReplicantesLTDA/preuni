package com.preuni.shared.domain.user

data class Student(
    val id: String,
    val displayName: String,
    val username: String,
    val email: String,
    val avatarUrl: String?,
    val xpTotal: Long,
    val streakCount: Int,
    val readinessScore: Double,
    val onboardingCompleted: Boolean,
)
