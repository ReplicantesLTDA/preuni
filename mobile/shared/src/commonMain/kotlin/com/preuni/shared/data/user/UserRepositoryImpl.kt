package com.preuni.shared.data.user

import com.preuni.shared.domain.user.AvatarUploadUrl
import com.preuni.shared.domain.user.Student
import com.preuni.shared.domain.user.UserRepository

class UserRepositoryImpl(private val apiClient: UserApiClient) : UserRepository {

    override suspend fun getMe(): Result<Student> = apiClient.getMe()

    override suspend fun updateProfile(displayName: String?, username: String?): Result<Student> =
        apiClient.updateProfile(displayName, username)

    override suspend fun getAvatarUploadUrl(): Result<AvatarUploadUrl> =
        apiClient.getAvatarUploadUrl()

    override suspend fun confirmAvatarUpload(objectKey: String): Result<Student> =
        apiClient.confirmAvatarUpload(objectKey)

    override suspend fun anonymize(): Result<Unit> = apiClient.anonymize()
}
