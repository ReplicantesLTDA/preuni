package com.preuni.shared.domain.error

/**
 * Typed error catalogue for all network and domain failures in the KMP shared module.
 * Maps to the error codes defined in the backend API contract.
 */
sealed class AppError(message: String) : Exception(message) {

    /** The device has no connectivity or the request could not be sent. */
    data class NetworkError(override val message: String = "No internet connection") : AppError(message)

    /** The access token is missing, invalid, or expired. The app should redirect to login. */
    data class Unauthorized(override val message: String = "Session expired. Please sign in again.") : AppError(message)

    /** The request was understood but refused (e.g., email not yet verified). */
    data class Forbidden(override val message: String = "Access denied.") : AppError(message)

    /** A conflict with existing state (e.g., email already registered). */
    data class Conflict(override val message: String) : AppError(message)

    /**
     * A field-level validation error returned by the server.
     * @param field The name of the field that failed validation.
     */
    data class Validation(val field: String, override val message: String) : AppError(message)

    /** A server-side error with no recovery path. */
    data class Unknown(override val message: String = "Something went wrong. Please try again.") : AppError(message)
}
