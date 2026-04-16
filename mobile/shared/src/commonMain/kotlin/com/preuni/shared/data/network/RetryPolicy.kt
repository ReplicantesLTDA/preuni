package com.preuni.shared.data.network

import com.preuni.shared.domain.error.AppError

/**
 * Encapsulates retry classification and backoff for Home/Profile fetch paths.
 *
 * Rules:
 * - Only [AppError.NetworkError] is transient; all other error types stop retries immediately.
 * - Backoff grows exponentially: 500ms, 1000ms, 2000ms (capped at 5s).
 * - Maximum 3 attempts total (initial + 2 retries).
 */
object RetryPolicy {

    /** Maximum number of total attempts (initial + retries). */
    const val MAX_ATTEMPTS = 3

    /**
     * Returns true when [error] represents a condition worth retrying automatically.
     * Network failures (connectivity issues, timeouts) are transient.
     * Auth, validation, and server errors are permanent and must not be retried.
     */
    fun isTransient(error: AppError): Boolean = when (error) {
        is AppError.NetworkError -> true
        is AppError.Unauthorized,
        is AppError.Forbidden,
        is AppError.Conflict,
        is AppError.Validation,
        is AppError.Unknown -> false
    }

    /**
     * Returns the delay in milliseconds before attempt [attempt] (0-indexed).
     * Attempt 0 → 500ms, 1 → 1000ms, 2 → 2000ms, then capped at 5000ms.
     */
    fun delayMillis(attempt: Int): Long = (500L * (1L shl attempt)).coerceAtMost(5_000L)
}
