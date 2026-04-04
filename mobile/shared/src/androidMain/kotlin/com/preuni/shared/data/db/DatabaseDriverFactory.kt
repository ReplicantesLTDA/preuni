package com.preuni.shared.data.db

import android.content.Context
import app.cash.sqldelight.db.SqlDriver
import app.cash.sqldelight.driver.android.AndroidSqliteDriver
import com.preuni.shared.data.db.PreuniDatabase

lateinit var appContext: Context

actual fun createSqlDriver(): SqlDriver =
    AndroidSqliteDriver(PreuniDatabase.Schema, appContext, "preuni.db")
