package com.preuni.shared.data.db

import app.cash.sqldelight.db.SqlDriver
import app.cash.sqldelight.driver.jdbc.sqlite.JdbcSqliteDriver

/** JVM/Desktop driver — uses an in-memory SQLite database. Used only in tests. */
actual fun createSqlDriver(): SqlDriver =
    JdbcSqliteDriver(JdbcSqliteDriver.IN_MEMORY).also { PreuniDatabase.Schema.create(it) }
