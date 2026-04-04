package com.preuni.shared.data.content

import com.preuni.shared.data.network.toAppError
import com.preuni.shared.domain.content.ContentRepository
import com.preuni.shared.domain.content.Track
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.http.isSuccess
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class TrackDto(
    val id: String,
    val name: String,
    val slug: String,
    @SerialName("color_token") val colorToken: String = "#6750A4",
    @SerialName("lesson_count") val lessonCount: Int = 0,
)

class ContentRepositoryImpl(private val httpClient: HttpClient) : ContentRepository {

    override suspend fun getTracks(): Result<List<Track>> = runCatching {
        val resp = httpClient.get("v1/tracks")
        if (!resp.status.isSuccess()) throw resp.toAppError()
        resp.body<List<TrackDto>>().map {
            Track(it.id, it.name, it.slug, it.colorToken, it.lessonCount)
        }
    }
}
