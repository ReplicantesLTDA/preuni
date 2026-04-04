package com.preuni.shared.data.user

import com.preuni.shared.data.network.toAppError
import com.preuni.shared.domain.user.AvatarUploadUrl
import com.preuni.shared.domain.user.Student
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.request.patch
import io.ktor.client.request.post
import io.ktor.client.request.put
import io.ktor.client.request.setBody
import io.ktor.http.isSuccess
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class StudentDto(
    val id: String,
    @SerialName("display_name") val displayName: String,
    val username: String,
    val email: String,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    @SerialName("xp_total") val xpTotal: Long = 0,
    @SerialName("streak_count") val streakCount: Int = 0,
    @SerialName("readiness_score") val readinessScore: Double = 0.0,
    @SerialName("onboarding_completed") val onboardingCompleted: Boolean = false,
)

@Serializable
data class UpdateProfileRequest(
    @SerialName("display_name") val displayName: String? = null,
    val username: String? = null,
)

@Serializable
data class AvatarUploadUrlDto(
    @SerialName("upload_url") val uploadUrl: String,
    @SerialName("object_key") val objectKey: String,
)

@Serializable
data class AvatarConfirmRequest(
    @SerialName("object_key") val objectKey: String,
)

class UserApiClient(private val httpClient: HttpClient) {

    suspend fun getMe(): Result<Student> = runCatching {
        val resp = httpClient.get("v1/students/me")
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<StudentDto>().toDomain()
    }

    suspend fun updateProfile(displayName: String?, username: String?): Result<Student> = runCatching {
        val resp = httpClient.patch("v1/students/me") {
            setBody(UpdateProfileRequest(displayName, username))
        }
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<StudentDto>().toDomain()
    }

    suspend fun getAvatarUploadUrl(): Result<AvatarUploadUrl> = runCatching {
        val resp = httpClient.put("v1/students/me/avatar")
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<AvatarUploadUrlDto>().let { AvatarUploadUrl(it.uploadUrl, it.objectKey) }
    }

    suspend fun confirmAvatarUpload(objectKey: String): Result<Student> = runCatching {
        val resp = httpClient.post("v1/students/me/avatar/confirm") {
            setBody(AvatarConfirmRequest(objectKey))
        }
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<StudentDto>().toDomain()
    }

    suspend fun anonymize(): Result<Unit> = runCatching {
        val resp = httpClient.delete("v1/students/me")
        if (!resp.status.isSuccess()) throw resp.toAppError()
    }
}

private fun StudentDto.toDomain() = Student(
    id = id,
    displayName = displayName,
    username = username,
    email = email,
    avatarUrl = avatarUrl,
    xpTotal = xpTotal,
    streakCount = streakCount,
    readinessScore = readinessScore,
    onboardingCompleted = onboardingCompleted,
)
