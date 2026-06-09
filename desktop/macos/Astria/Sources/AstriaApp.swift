import SwiftUI
import WebKit

private enum AstriaDefaults {
    static let webURL = "http://127.0.0.1:7533/app/"
    static let diagnosticsURL = "http://127.0.0.1:7533/diagnostics"
}

@main
struct AstriaApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate

    var body: some Scene {
        WindowGroup("Astria") {
            AstriaRootView(config: LaunchConfig.fromProcess())
                .frame(minWidth: 1040, minHeight: 720)
        }
        .commands {
            CommandGroup(replacing: .newItem) {}
        }
    }
}

final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        true
    }
}

struct LaunchConfig {
    let webURL: URL
    let diagnosticsURL: URL

    static func fromProcess(
        arguments: [String] = CommandLine.arguments,
        environment: [String: String] = ProcessInfo.processInfo.environment
    ) -> LaunchConfig {
        let webURLString = firstValue(after: "--web-url", in: arguments)
            ?? environment["ASTRIA_WEB_URL"]
            ?? AstriaDefaults.webURL
        let diagnosticsURLString = firstValue(after: "--diagnostics-url", in: arguments)
            ?? environment["ASTRIA_DIAGNOSTICS_URL"]
            ?? AstriaDefaults.diagnosticsURL

        return LaunchConfig(
            webURL: URL(string: webURLString) ?? URL(string: AstriaDefaults.webURL)!,
            diagnosticsURL: URL(string: diagnosticsURLString) ?? URL(string: AstriaDefaults.diagnosticsURL)!
        )
    }

    private static func firstValue(after flag: String, in arguments: [String]) -> String? {
        guard let index = arguments.firstIndex(of: flag) else {
            return nil
        }
        let valueIndex = arguments.index(after: index)
        guard valueIndex < arguments.endIndex else {
            return nil
        }
        return arguments[valueIndex]
    }
}

struct AstriaRootView: View {
    let config: LaunchConfig
    @State private var loadState = WebLoadState()

    var body: some View {
        ZStack(alignment: .top) {
            AstriaWebView(url: config.webURL, loadState: $loadState)
                .ignoresSafeArea()
            if let message = loadState.message {
                StatusBanner(message: message, diagnosticsURL: config.diagnosticsURL)
                    .padding(.top, 12)
                    .padding(.horizontal, 16)
            }
        }
    }
}

struct StatusBanner: View {
    let message: String
    let diagnosticsURL: URL

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: "exclamationmark.triangle")
                .imageScale(.medium)
            Text(message)
                .lineLimit(2)
            Spacer(minLength: 16)
            Button("Diagnostics") {
                NSWorkspace.shared.open(diagnosticsURL)
            }
        }
        .font(.system(size: 13))
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
        .shadow(radius: 8, y: 3)
    }
}

struct WebLoadState {
    var message: String?
}

struct AstriaWebView: NSViewRepresentable {
    let url: URL
    @Binding var loadState: WebLoadState

    func makeCoordinator() -> Coordinator {
        Coordinator(loadState: $loadState)
    }

    func makeNSView(context: Context) -> WKWebView {
        let configuration = WKWebViewConfiguration()
        configuration.websiteDataStore = .default()

        let view = WKWebView(frame: .zero, configuration: configuration)
        view.navigationDelegate = context.coordinator
        view.allowsBackForwardNavigationGestures = true
        view.load(URLRequest(url: url))
        return view
    }

    func updateNSView(_ view: WKWebView, context: Context) {
        if view.url == nil {
            view.load(URLRequest(url: url))
        }
    }

    final class Coordinator: NSObject, WKNavigationDelegate {
        @Binding private var loadState: WebLoadState

        init(loadState: Binding<WebLoadState>) {
            _loadState = loadState
        }

        func webView(_ webView: WKWebView, didStartProvisionalNavigation navigation: WKNavigation!) {
            loadState.message = nil
        }

        func webView(_ webView: WKWebView, didFinish navigation: WKNavigation!) {
            loadState.message = nil
        }

        func webView(_ webView: WKWebView, didFail navigation: WKNavigation!, withError error: Error) {
            loadState.message = "Astria could not finish loading. Start StarClaw with `starclaw app --no-open`, then reload this window."
        }

        func webView(_ webView: WKWebView, didFailProvisionalNavigation navigation: WKNavigation!, withError error: Error) {
            loadState.message = "Astria is waiting for the local daemon at \(AstriaDefaults.webURL)."
        }
    }
}
