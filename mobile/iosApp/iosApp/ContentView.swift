import SwiftUI
import Shared

/// Root SwiftUI view that hosts the Kotlin Multiplatform Compose UI.
///
/// The KMP shared module is built as an XCFramework by running:
///   ./gradlew :shared:assembleXCFramework
/// which produces `mobile/build/XCFrameworks/release/Shared.xcframework`.
/// That framework must be added to the Xcode target's "Frameworks, Libraries,
/// and Embedded Content" section (Embed & Sign).
struct ContentView: View {
    var body: some View {
        ComposeView()
            .ignoresSafeArea(.keyboard)
    }
}

/// UIViewControllerRepresentable that bridges to MainViewController from KMP.
struct ComposeView: UIViewControllerRepresentable {
    func makeUIViewController(context: Context) -> UIViewController {
        MainViewControllerKt.MainViewController()
    }

    func updateUIViewController(_ uiViewController: UIViewController, context: Context) {}
}
