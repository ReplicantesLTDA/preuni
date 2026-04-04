package com.preuni.shared.domain.auth

data class AuthSession(
    val userId: String,
    val userEmail: String,
    val accessToken: String,
    val refreshToken: String,
)
