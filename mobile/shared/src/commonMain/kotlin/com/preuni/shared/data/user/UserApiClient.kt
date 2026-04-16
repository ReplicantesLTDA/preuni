package com.preuni.shared.data.user

import com.preuni.shared.data.network.toAppError
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.user.AvatarUploadUrl
import com.preuni.shared.domain.user.Student
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.plugins.timeout
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
        val resp = httpClient.get("v1/students/me") {
            timeout { requestTimeoutMillis = 10_000 }
        }
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<StudentDto>().validated().toDomain()
    }.mapToAppError()

    suspend fun updateTracks(trackIds: List<String>): Result<Unit> = runCatching {
        val resp = httpClient.patch("v1/students/me/onboarding") {
            setBody(mapOf("enrolled_track_ids" to trackIds))
        }
        if (!resp.status.isSuccess()) throw resp.toAppError()
    }.mapToAppError()

    suspend fun updateProfile(displayName: String?, username: String?): Result<Student> = runCatching {
        val resp = httpClient.patch("v1/students/me") {
            timeout { requestTimeoutMillis = 10_000 }
            setBody(UpdateProfileRequest(displayName, username))
        }
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<StudentDto>().validated().toDomain()
    }.mapToAppError()

    suspend fun getAvatarUploadUrl(): Result<AvatarUploadUrl> = runCatching {
        val resp = httpClient.put("v1/students/me/avatar")
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<AvatarUploadUrlDto>().let { AvatarUploadUrl(it.uploadUrl, it.objectKey) }
    }.mapToAppError()

    suspend fun confirmAvatarUpload(objectKey: String): Result<Student> = runCatching {
        val resp = httpClient.post("v1/students/me/avatar/confirm") {
            setBody(AvatarConfirmRequest(objectKey))
        }
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<StudentDto>().validated().toDomain()
    }.mapToAppError()

    suspend fun anonymize(): Result<Unit> = runCatching {
        val resp = httpClient.delete("v1/students/me")
        if (!resp.status.isSuccess()) throw resp.toAppError()
    }.mapToAppError()
}

/**
 * Validates required fields in the API response before domain mapping.
 * Throws [IllegalArgumentException] on contract violations — callers map this to
 * [AppError.Unknown] via [mapToAppError].
 */
private fun StudentDto.validated(): StudentDto {
    require(id.isNotBlank()) { "StudentResponse: id must not be blank" }
    require(displayName.isNotBlank()) { "StudentResponse: display_name must not be blank" }
    require(username.isNotBlank()) { "StudentResponse: username must not be blank" }
    require(email.isNotBlank()) { "StudentResponse: email must not be blank" }
    require(xpTotal >= 0) { "StudentResponse: xp_total must be >= 0" }
    require(streakCount >= 0) { "StudentResponse: streak_count must be >= 0" }
    require(readinessScore in 0.0..1.0) { "StudentResponse: readiness_score must be in [0.0, 1.0]" }
    return this
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

/**
 * Maps any non-[AppError] exception to a typed [AppError]:
 * - [IllegalArgumentException] → [AppError.Unknown] (malformed server response)
 * - Everything else → [AppError.NetworkError] (connectivity / timeout)
 */
private fun <T> Result<T>.mapToAppError(): Result<T> = fold(
    onSuccess = { Result.success(it) },
    onFailure = { e ->
        Result.failure(
            when (e) {
                is AppError -> e
                is IllegalArgumentException -> AppError.Unknown(
                    "Malformed server response: ${e.message}"
                )
                else -> AppError.NetworkError(e.message ?: "Network error")
            }
        )
    },
)

