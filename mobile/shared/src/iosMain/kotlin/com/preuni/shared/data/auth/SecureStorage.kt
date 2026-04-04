package com.preuni.shared.data.auth

import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.alloc
import kotlinx.cinterop.memScoped
import kotlinx.cinterop.ptr
import kotlinx.cinterop.value
import platform.CoreFoundation.CFDictionaryCreateMutable
import platform.CoreFoundation.CFDictionarySetValue
import platform.CoreFoundation.CFTypeRefVar
import platform.CoreFoundation.kCFAllocatorDefault
import platform.Foundation.NSData
import platform.Foundation.NSString
import platform.Foundation.NSUTF8StringEncoding
import platform.Foundation.dataUsingEncoding
import platform.Foundation.create
import platform.Security.SecItemAdd
import platform.Security.SecItemCopyMatching
import platform.Security.SecItemDelete
import platform.Security.SecItemUpdate
import platform.Security.errSecSuccess
import platform.Security.kSecAttrAccount
import platform.Security.kSecAttrService
import platform.Security.kSecClass
import platform.Security.kSecClassGenericPassword
import platform.Security.kSecMatchLimit
import platform.Security.kSecMatchLimitOne
import platform.Security.kSecReturnData
import platform.Security.kSecValueData

@OptIn(ExperimentalForeignApi::class)
actual class SecureStorage actual constructor() {

    private val service = "com.preuni.app"

    private fun buildQuery(key: String) = mapOf<Any?, Any?>(
        kSecClass to kSecClassGenericPassword,
        kSecAttrService to service,
        kSecAttrAccount to key,
    )

    actual fun put(key: String, value: String) {
        val data = (value as NSString).dataUsingEncoding(NSUTF8StringEncoding) ?: return
        val query = buildQuery(key)
        val status = SecItemUpdate(query, mapOf(kSecValueData to data))
        if (status != errSecSuccess) {
            SecItemAdd(query + mapOf(kSecValueData to data), null)
        }
    }

    actual fun get(key: String): String? {
        val query = buildQuery(key) + mapOf<Any?, Any?>(
            kSecReturnData to true,
            kSecMatchLimit to kSecMatchLimitOne,
        )
        memScoped {
            val result = alloc<CFTypeRefVar>()
            val status = SecItemCopyMatching(query, result.ptr)
            if (status != errSecSuccess) return null
            val data = result.value as? NSData ?: return null
            return NSString.create(data, NSUTF8StringEncoding) as? String
        }
    }

    actual fun remove(key: String) {
        SecItemDelete(buildQuery(key))
    }

    actual fun clear() {
        SecItemDelete(
            mapOf<Any?, Any?>(
                kSecClass to kSecClassGenericPassword,
                kSecAttrService to service,
            )
        )
    }
}
