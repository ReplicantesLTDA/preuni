package com.preuni.shared.domain.content

interface ContentRepository {
    suspend fun getTracks(): Result<List<Track>>
}

data class Track(
    val id: String,
    val name: String,
    val slug: String,
    val colorToken: String,
    val lessonCount: Int,
)
