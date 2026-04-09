package com.preuni.android

import android.app.Application
import com.preuni.shared.data.db.appContext

class PreuniApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        appContext = this
    }
}
