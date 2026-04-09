package com.preuni.shared.data.auth

import cnames.structs.__CFData
import cnames.structs.__CFDictionary
import kotlinx.cinterop.CPointer
import kotlinx.cinterop.CValuesRef
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.alloc
import kotlinx.cinterop.get
import kotlinx.cinterop.memScoped
import kotlinx.cinterop.ptr
import kotlinx.cinterop.reinterpret
import kotlinx.cinterop.toCValues
import kotlinx.cinterop.value
import platform.CoreFoundation.CFDataCreate
import platform.CoreFoundation.CFDataGetBytePtr
import platform.CoreFoundation.CFDataGetLength
import platform.CoreFoundation.CFDictionaryCreateMutable
import platform.CoreFoundation.CFDictionarySetValue
import platform.CoreFoundation.CFRelease
import platform.CoreFoundation.CFStringCreateWithCString
import platform.CoreFoundation.CFTypeRefVar
import platform.CoreFoundation.kCFAllocatorDefault
import platform.CoreFoundation.kCFBooleanTrue
import platform.CoreFoundation.kCFStringEncodingUTF8
import platform.CoreFoundation.kCFTypeDictionaryKeyCallBacks
import platform.CoreFoundation.kCFTypeDictionaryValueCallBacks
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

    private fun cfString(value: String): CPointer<cnames.structs.__CFString>? {
        return CFStringCreateWithCString(kCFAllocatorDefault, value, kCFStringEncodingUTF8)
    }

    private fun createDictionary(vararg entries: Pair<CValuesRef<*>?, CValuesRef<*>?>): CPointer<__CFDictionary>? {
        val dictionary = CFDictionaryCreateMutable(
            kCFAllocatorDefault,
            entries.size.toLong(),
            kCFTypeDictionaryKeyCallBacks.ptr,
            kCFTypeDictionaryValueCallBacks.ptr,
        ) ?: return null

        entries.forEach { (key, value) ->
            CFDictionarySetValue(dictionary, key, value)
        }

        return dictionary
    }

    actual fun put(key: String, value: String) {
        val serviceRef = cfString(service) ?: return
        val accountRef = cfString(key)
        if (accountRef == null) {
            CFRelease(serviceRef)
            return
        }

        val valueBytes = value.encodeToByteArray().toUByteArray()
        val dataRef = CFDataCreate(kCFAllocatorDefault, valueBytes.toCValues(), valueBytes.size.toLong())
        if (dataRef == null) {
            CFRelease(serviceRef)
            CFRelease(accountRef)
            return
        }

        val query = createDictionary(
            kSecClass to kSecClassGenericPassword,
            kSecAttrService to serviceRef,
            kSecAttrAccount to accountRef,
        )
        val attributes = createDictionary(kSecValueData to dataRef)

        if (query != null && attributes != null) {
            val status = SecItemUpdate(query, attributes)
            if (status != errSecSuccess) {
                val addAttributes = createDictionary(
                    kSecClass to kSecClassGenericPassword,
                    kSecAttrService to serviceRef,
                    kSecAttrAccount to accountRef,
                    kSecValueData to dataRef,
                )
                if (addAttributes != null) {
                    SecItemAdd(addAttributes, null)
                    CFRelease(addAttributes)
                }
            }
            CFRelease(query)
            CFRelease(attributes)
        }

        CFRelease(dataRef)
        CFRelease(serviceRef)
        CFRelease(accountRef)
    }

    actual fun get(key: String): String? {
        val serviceRef = cfString(service) ?: return null
        val accountRef = cfString(key)
        if (accountRef == null) {
            CFRelease(serviceRef)
            return null
        }

        val query = createDictionary(
            kSecClass to kSecClassGenericPassword,
            kSecAttrService to serviceRef,
            kSecAttrAccount to accountRef,
            kSecReturnData to kCFBooleanTrue,
            kSecMatchLimit to kSecMatchLimitOne,
        )

        CFRelease(serviceRef)
        CFRelease(accountRef)

        if (query == null) return null

        memScoped {
            val result = alloc<CFTypeRefVar>()
            val status = SecItemCopyMatching(query, result.ptr)
            CFRelease(query)

            if (status != errSecSuccess) return null

            val resultRef = result.value ?: return null
            try {
                val dataRef = resultRef.reinterpret<__CFData>()
                val length = CFDataGetLength(dataRef).toInt()
                val bytesPtr = CFDataGetBytePtr(dataRef) ?: return null
                val bytes = ByteArray(length) { index -> bytesPtr[index].toByte() }
                return bytes.decodeToString()
            } finally {
                CFRelease(resultRef)
            }
        }
    }

    actual fun remove(key: String) {
        val serviceRef = cfString(service) ?: return
        val accountRef = cfString(key)
        if (accountRef == null) {
            CFRelease(serviceRef)
            return
        }

        val query = createDictionary(
            kSecClass to kSecClassGenericPassword,
            kSecAttrService to serviceRef,
            kSecAttrAccount to accountRef,
        )

        CFRelease(serviceRef)
        CFRelease(accountRef)

        if (query != null) {
            SecItemDelete(query)
            CFRelease(query)
        }
    }

    actual fun clear() {
        val serviceRef = cfString(service) ?: return
        val query = createDictionary(
            kSecClass to kSecClassGenericPassword,
            kSecAttrService to serviceRef,
        )

        CFRelease(serviceRef)

        if (query != null) {
            SecItemDelete(query)
            CFRelease(query)
        }
    }
}
