package com.preuni.shared.domain.auth

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue

class AuthValidatorTest {

    // ── Email ─────────────────────────────────────────────────────────────────

    @Test
    fun `validateEmail - valid email returns Valid`() {
        val result = AuthValidator.validateEmail("student@example.com")
        assertIs<ValidationResult.Valid>(result)
    }

    @Test
    fun `validateEmail - missing at-sign returns Invalid`() {
        val result = AuthValidator.validateEmail("noatsign.com")
        assertIs<ValidationResult.Invalid>(result)
    }

    @Test
    fun `validateEmail - empty string returns Invalid`() {
        val result = AuthValidator.validateEmail("")
        assertIs<ValidationResult.Invalid>(result)
        assertEquals("Email is required", (result as ValidationResult.Invalid).message)
    }

    @Test
    fun `validateEmail - missing domain returns Invalid`() {
        val result = AuthValidator.validateEmail("user@")
        assertIs<ValidationResult.Invalid>(result)
    }

    // ── Username ──────────────────────────────────────────────────────────────

    @Test
    fun `validateUsername - valid username with letters and digits returns Valid`() {
        assertTrue(AuthValidator.validateUsername("ana01").isValid)
    }

    @Test
    fun `validateUsername - valid username with hyphen and underscore returns Valid`() {
        assertTrue(AuthValidator.validateUsername("jo-se_01").isValid)
    }

    @Test
    fun `validateUsername - uppercase letter returns Invalid`() {
        val result = AuthValidator.validateUsername("AnaUser")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("lowercase"))
    }

    @Test
    fun `validateUsername - starts with digit returns Invalid`() {
        val result = AuthValidator.validateUsername("1user")
        assertIs<ValidationResult.Invalid>(result)
    }

    @Test
    fun `validateUsername - contains forward slash returns Invalid`() {
        val result = AuthValidator.validateUsername("user/name")
        assertIs<ValidationResult.Invalid>(result)
    }

    @Test
    fun `validateUsername - too short (2 chars) returns Invalid`() {
        val result = AuthValidator.validateUsername("ab")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("3 characters"))
    }

    @Test
    fun `validateUsername - too long (31 chars) returns Invalid`() {
        val result = AuthValidator.validateUsername("a" + "b".repeat(30)) // 31 chars
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("30 characters"))
    }

    @Test
    fun `validateUsername - exactly 3 chars returns Valid`() {
        assertTrue(AuthValidator.validateUsername("abc").isValid)
    }

    @Test
    fun `validateUsername - exactly 30 chars returns Valid`() {
        val name = "a" + "b".repeat(28) + "c" // 30 chars
        assertTrue(AuthValidator.validateUsername(name).isValid)
    }

    // ── Password ──────────────────────────────────────────────────────────────

    @Test
    fun `validatePassword - valid password returns Valid`() {
        assertTrue(AuthValidator.validatePassword("Abc123").isValid)
    }

    @Test
    fun `validatePassword - too short (5 chars) returns Invalid`() {
        val result = AuthValidator.validatePassword("Ab1cd")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("6 characters"))
    }

    @Test
    fun `validatePassword - exactly 6 chars returns Valid`() {
        assertTrue(AuthValidator.validatePassword("Abc12d").isValid)
    }

    @Test
    fun `validatePassword - too long (256 chars) returns Invalid`() {
        val long = "Aa1" + "x".repeat(253) // 256 chars
        val result = AuthValidator.validatePassword(long)
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("255 characters"))
    }

    @Test
    fun `validatePassword - exactly 255 chars returns Valid`() {
        val p = "Aa1" + "x".repeat(252) // 255 chars
        assertTrue(AuthValidator.validatePassword(p).isValid)
    }

    @Test
    fun `validatePassword - no uppercase returns Invalid`() {
        val result = AuthValidator.validatePassword("abc123def")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("uppercase"))
    }

    @Test
    fun `validatePassword - no lowercase returns Invalid`() {
        val result = AuthValidator.validatePassword("ABC123DEF")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("lowercase"))
    }

    @Test
    fun `validatePassword - no digit returns Invalid`() {
        val result = AuthValidator.validatePassword("AbcDefGhi")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("digit"))
    }

    // ── Passwords match ───────────────────────────────────────────────────────

    @Test
    fun `validatePasswordsMatch - matching passwords returns Valid`() {
        assertTrue(AuthValidator.validatePasswordsMatch("Abc123", "Abc123").isValid)
    }

    @Test
    fun `validatePasswordsMatch - different passwords returns Invalid`() {
        val result = AuthValidator.validatePasswordsMatch("Abc123", "Abc124")
        assertIs<ValidationResult.Invalid>(result)
        assertTrue(result.message.contains("not match"))
    }
}
