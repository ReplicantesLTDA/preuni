package com.preuni.shared.domain.user

interface UserRepository {
    suspend fun getMe(): Result<Student>
    suspend fun updateProfile(displayName: String? = null, username: String? = null): Result<Student>
    suspend fun getAvatarUploadUrl(): Result<AvatarUploadUrl>
    suspend fun confirmAvatarUpload(objectKey: String): Result<Student>
    suspend fun anonymize(): Result<Unit>
}

data class AvatarUploadUrl(
    val uploadUrl: String,
    val objectKey: String,
)
