package com.preuni.shared.domain.auth

/**
 * Result of a field-level validation check.
 */
sealed class ValidationResult {
    /** The value is valid and can be submitted. */
    data object Valid : ValidationResult()

    /** The value is invalid. [message] is suitable for display in the UI. */
    data class Invalid(val message: String) : ValidationResult()
}

/** True when this result represents a passing validation. */
val ValidationResult.isValid: Boolean get() = this is ValidationResult.Valid

/**
 * Stateless validator for all authentication-related form fields.
 *
 * Rules:
 * - **Email**: must match basic RFC 5322 format (`local@domain.tld`).
 * - **Username**: 3–30 characters; lowercase letters, digits, hyphen and underscore only;
 *   must start with a letter; must end with a letter or digit.
 * - **Password**: 6–255 characters; at least one uppercase letter, one lowercase letter,
 *   and one digit.
 *
 * Note: "slash" in the original spec has been interpreted as underscore (`_`).
 * If forward-slash `/` should be allowed, add it to [USERNAME_ALLOWED_CHARS_REGEX].
 */
object AuthValidator {

    private val EMAIL_REGEX = Regex("""^[^@\s]+@[^@\s]+\.[^@\s]+$""")

    // 3–30 chars: starts with [a-z], middle is [a-z0-9_-]*, ends with [a-z0-9]
    private val USERNAME_REGEX = Regex("""^[a-z][a-z0-9_-]{1,28}[a-z0-9]$""")

    private const val PASSWORD_MIN = 6
    private const val PASSWORD_MAX = 255

    /** Validates an email address. */
    fun validateEmail(email: String): ValidationResult {
        val trimmed = email.trim()
        if (trimmed.isEmpty()) return ValidationResult.Invalid("Email is required")
        if (!EMAIL_REGEX.matches(trimmed)) return ValidationResult.Invalid("Enter a valid email address")
        return ValidationResult.Valid
    }

    /**
     * Validates a username.
     * Allowed: lowercase letters (a–z), digits (0–9), hyphen (-), underscore (_).
     * Must start with a letter; must end with a letter or digit.
     * Length: 3–30 characters.
     */
    fun validateUsername(username: String): ValidationResult {
        if (username.isEmpty()) return ValidationResult.Invalid("Username is required")
        if (username.length < 3) return ValidationResult.Invalid("Username must be at least 3 characters")
        if (username.length > 30) return ValidationResult.Invalid("Username must be at most 30 characters")
        if (!USERNAME_REGEX.matches(username)) {
            return ValidationResult.Invalid(
                "Username may only contain lowercase letters, digits, - and _; must start with a letter"
            )
        }
        return ValidationResult.Valid
    }

    /**
     * Validates a password.
     * Rules: 6–255 characters, at least one uppercase letter, one lowercase letter, one digit.
     */
    fun validatePassword(password: String): ValidationResult {
        if (password.length < PASSWORD_MIN) return ValidationResult.Invalid("Password must be at least 6 characters")
        if (password.length > PASSWORD_MAX) return ValidationResult.Invalid("Password must be at most 255 characters")
        if (!password.any { it.isUpperCase() }) return ValidationResult.Invalid("Password must contain at least one uppercase letter")
        if (!password.any { it.isLowerCase() }) return ValidationResult.Invalid("Password must contain at least one lowercase letter")
        if (!password.any { it.isDigit() }) return ValidationResult.Invalid("Password must contain at least one digit")
        return ValidationResult.Valid
    }

    /**
     * Validates that two password fields match (for confirm-password fields).
     */
    fun validatePasswordsMatch(password: String, confirm: String): ValidationResult {
        if (password != confirm) return ValidationResult.Invalid("Passwords do not match")
        return ValidationResult.Valid
    }
}
