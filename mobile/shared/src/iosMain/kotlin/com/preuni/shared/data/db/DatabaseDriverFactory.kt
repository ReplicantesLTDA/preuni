package com.preuni.shared.data.db

import app.cash.sqldelight.db.SqlDriver
import app.cash.sqldelight.driver.native.NativeSqliteDriver
import com.preuni.shared.data.db.PreuniDatabase

actual fun createSqlDriver(): SqlDriver =
    NativeSqliteDriver(PreuniDatabase.Schema, "preuni.db")
