package com.preuni.shared.data.db

import app.cash.sqldelight.db.SqlDriver
import app.cash.sqldelight.driver.worker.WebWorkerDriver
import org.w3c.dom.Worker

private fun sqlJsWorkerUrl(): String =
    js("new URL('@cashapp/sqldelight-sqljs-worker/sqljs.worker.js', import.meta.url).toString()")

actual fun createSqlDriver(): SqlDriver =
    WebWorkerDriver(
        Worker(sqlJsWorkerUrl())
    )
