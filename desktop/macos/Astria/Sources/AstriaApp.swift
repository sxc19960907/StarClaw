import SwiftUI
import WebKit

private enum AstriaDefaults {
    static let webURL = "http://127.0.0.1:7533/app/"
    static let healthURL = "http://127.0.0.1:7533/health"
    static let diagnosticsURL = "http://127.0.0.1:7533/diagnostics"
    static let startupTimeoutSeconds: TimeInterval = 5
    static let healthProbeTimeoutSeconds: TimeInterval = 0.5
    static let healthPollIntervalSeconds: TimeInterval = 0.12
}

@main
struct AstriaApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate

    init() {
        if CommandLine.arguments.contains("--supervision-smoke") {
            Foundation.exit(AstriaSupervisorSmoke.run())
        }
    }

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
    let healthURL: URL
    let diagnosticsURL: URL
    let starclawBinary: String?
    let startupTimeout: TimeInterval
    let healthProbeTimeout: TimeInterval
    let healthPollInterval: TimeInterval

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
        let healthURLString = firstValue(after: "--health-url", in: arguments)
            ?? environment["ASTRIA_HEALTH_URL"]
            ?? AstriaDefaults.healthURL
        let startupTimeout = timeIntervalValue(
            firstValue(after: "--startup-timeout", in: arguments)
                ?? environment["ASTRIA_STARTUP_TIMEOUT"],
            defaultValue: AstriaDefaults.startupTimeoutSeconds
        )

        return LaunchConfig(
            webURL: URL(string: webURLString) ?? URL(string: AstriaDefaults.webURL)!,
            healthURL: URL(string: healthURLString) ?? URL(string: AstriaDefaults.healthURL)!,
            diagnosticsURL: URL(string: diagnosticsURLString) ?? URL(string: AstriaDefaults.diagnosticsURL)!,
            starclawBinary: firstValue(after: "--starclaw-bin", in: arguments)
                ?? environment["ASTRIA_STARCLAW_BIN"],
            startupTimeout: startupTimeout,
            healthProbeTimeout: AstriaDefaults.healthProbeTimeoutSeconds,
            healthPollInterval: AstriaDefaults.healthPollIntervalSeconds
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

    private static func timeIntervalValue(_ value: String?, defaultValue: TimeInterval) -> TimeInterval {
        guard let value, let parsed = TimeInterval(value), parsed > 0 else {
            return defaultValue
        }
        return parsed
    }
}

struct AstriaRootView: View {
    let config: LaunchConfig
    @State private var loadState = WebLoadState()
    @StateObject private var supervisor: DaemonSupervisor

    init(config: LaunchConfig) {
        self.config = config
        _supervisor = StateObject(wrappedValue: DaemonSupervisor(config: config))
    }

    var body: some View {
        ZStack(alignment: .top) {
            if supervisor.state.isAttached {
                AstriaWebView(url: config.webURL, loadState: $loadState)
                    .ignoresSafeArea()
            } else {
                LaunchStateView(state: supervisor.state, diagnosticsURL: config.diagnosticsURL) {
                    supervisor.start()
                }
            }
            if let message = loadState.message ?? supervisor.state.bannerMessage {
                StatusBanner(message: message, diagnosticsURL: config.diagnosticsURL, canRetry: supervisor.state.canRetry) {
                    supervisor.start()
                }
                    .padding(.top, 12)
                    .padding(.horizontal, 16)
            }
        }
        .task {
            supervisor.start()
        }
    }
}

struct StatusBanner: View {
    let message: String
    let diagnosticsURL: URL
    let canRetry: Bool
    let retry: () -> Void

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
            if canRetry {
                Button("Retry") {
                    retry()
                }
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

struct LaunchStateView: View {
    let state: DaemonState
    let diagnosticsURL: URL
    let retry: () -> Void

    var body: some View {
        VStack(spacing: 18) {
            ProgressView()
                .controlSize(.large)
                .opacity(state.showsProgress ? 1 : 0)
            Text(state.title)
                .font(.system(size: 24, weight: .semibold))
            Text(state.detail)
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
                .frame(maxWidth: 560)
            HStack {
                Button("Diagnostics") {
                    NSWorkspace.shared.open(diagnosticsURL)
                }
                if state.canRetry {
                    Button("Retry") {
                        retry()
                    }
                    .keyboardShortcut(.defaultAction)
                }
            }
        }
        .padding(40)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct WebLoadState {
    var message: String?
}

enum DaemonState: Equatable {
    case idle
    case checking
    case starting
    case attached
    case failed(String)
    case crashed(String)

    var isAttached: Bool {
        if case .attached = self {
            return true
        }
        return false
    }

    var canRetry: Bool {
        switch self {
        case .failed, .crashed:
            return true
        default:
            return false
        }
    }

    var showsProgress: Bool {
        switch self {
        case .idle, .checking, .starting:
            return true
        default:
            return false
        }
    }

    var title: String {
        switch self {
        case .idle, .checking:
            return "Checking local daemon"
        case .starting:
            return "Starting StarClaw daemon"
        case .attached:
            return "Opening Astria"
        case .failed:
            return "Astria could not start the daemon"
        case .crashed:
            return "StarClaw daemon stopped"
        }
    }

    var detail: String {
        switch self {
        case .idle:
            return "Preparing the local Astria workspace."
        case .checking:
            return "Looking for an existing StarClaw daemon on this machine."
        case .starting:
            return "No healthy daemon was found, so Astria is starting one locally."
        case .attached:
            return "The local daemon is healthy."
        case .failed(let message), .crashed(let message):
            return message
        }
    }

    var bannerMessage: String? {
        switch self {
        case .checking:
            return "Astria is checking the local daemon."
        case .starting:
            return "Astria is starting the local daemon."
        case .failed(let message), .crashed(let message):
            return message
        default:
            return nil
        }
    }
}

@MainActor
final class DaemonSupervisor: ObservableObject {
    @Published private(set) var state: DaemonState = .idle

    private let config: LaunchConfig
    private var childProcess: Process?
    private var launchTask: Task<Void, Never>?

    init(config: LaunchConfig) {
        self.config = config
    }

    func start() {
        launchTask?.cancel()
        launchTask = Task { [weak self] in
            guard let self else {
                return
            }
            await self.supervise()
        }
    }

    private func supervise() async {
        state = .checking
        if await isHealthy() {
            state = .attached
            return
        }

        state = .starting
        let process: Process
        do {
            process = try launchDaemon()
            childProcess = process
        } catch {
            state = .failed("Could not start `starclaw daemon start`: \(error.localizedDescription)")
            return
        }

        let becameHealthy = await waitForHealth()
        if becameHealthy {
            state = .attached
            return
        }

        let message = process.isRunning
            ? "Daemon did not become healthy within \(Int(config.startupTimeout)) seconds. Check whether port 7533 is already in use or open diagnostics."
            : "Daemon process exited before becoming healthy. Open diagnostics or run `starclaw app --check` in Terminal."
        state = .failed(message)
    }

    private func launchDaemon() throws -> Process {
        try DaemonLaunchSupport.launchDaemon(config: config) { [weak self] proc in
            Task { @MainActor [weak self] in
                self?.handleDaemonExit(status: proc.terminationStatus)
            }
        }
    }

    private func handleDaemonExit(status: Int32) {
        childProcess = nil
        if state == .attached {
            state = .crashed("The daemon process started by Astria exited with status \(status). Retry to start it again.")
        }
    }

    private func waitForHealth() async -> Bool {
        let deadline = Date().addingTimeInterval(config.startupTimeout)
        while Date() < deadline {
            if Task.isCancelled {
                return false
            }
            if await isHealthy() {
                return true
            }
            try? await Task.sleep(nanoseconds: UInt64(config.healthPollInterval * 1_000_000_000))
        }
        return false
    }

    private func isHealthy() async -> Bool {
        var request = URLRequest(url: config.healthURL)
        request.timeoutInterval = config.healthProbeTimeout
        do {
            let response = try await URLSession.shared.data(for: request).1
            return (response as? HTTPURLResponse)?.statusCode == 200
        } catch {
            return false
        }
    }
}

enum DaemonLaunchSupport {
    static func launchDaemon(config: LaunchConfig, terminationHandler: ((Process) -> Void)? = nil) throws -> Process {
        let process = Process()
        let resolved = resolveStarclawBinary(config: config)
        process.executableURL = resolved.executableURL
        process.arguments = resolved.arguments + ["daemon", "start"]

        let devNull = FileHandle(forWritingAtPath: "/dev/null")
        process.standardOutput = devNull
        process.standardError = devNull
        process.terminationHandler = terminationHandler
        try process.run()
        return process
    }

    static func resolveStarclawBinary(config: LaunchConfig) -> (executableURL: URL, arguments: [String]) {
        if let override = config.starclawBinary, !override.isEmpty {
            return (URL(fileURLWithPath: override), [])
        }

        let bundled = Bundle.main.resourceURL?.appendingPathComponent("starclaw").path
        if let bundled, FileManager.default.isExecutableFile(atPath: bundled) {
            return (URL(fileURLWithPath: bundled), [])
        }

        return (URL(fileURLWithPath: "/usr/bin/env"), ["starclaw"])
    }

    static func isHealthy(url: URL, timeout: TimeInterval) -> Bool {
        var request = URLRequest(url: url)
        request.timeoutInterval = timeout

        let semaphore = DispatchSemaphore(value: 0)
        var healthy = false
        let task = URLSession.shared.dataTask(with: request) { _, response, _ in
            healthy = (response as? HTTPURLResponse)?.statusCode == 200
            semaphore.signal()
        }
        task.resume()
        _ = semaphore.wait(timeout: .now() + timeout + 0.25)
        task.cancel()
        return healthy
    }

    static func waitForHealth(config: LaunchConfig) -> Bool {
        let deadline = Date().addingTimeInterval(config.startupTimeout)
        while Date() < deadline {
            if isHealthy(url: config.healthURL, timeout: config.healthProbeTimeout) {
                return true
            }
            Thread.sleep(forTimeInterval: config.healthPollInterval)
        }
        return false
    }
}

enum AstriaSupervisorSmoke {
    static func run() -> Int32 {
        let config = LaunchConfig.fromProcess()

        if DaemonLaunchSupport.isHealthy(url: config.healthURL, timeout: config.healthProbeTimeout) {
            return 0
        }

        let process: Process
        do {
            process = try DaemonLaunchSupport.launchDaemon(config: config)
        } catch {
            fputs("Astria supervision smoke failed to launch daemon: \(error.localizedDescription)\n", stderr)
            return 1
        }

        if DaemonLaunchSupport.waitForHealth(config: config) {
            return 0
        }

        if process.isRunning {
            fputs("Astria supervision smoke timed out waiting for daemon health\n", stderr)
        } else {
            fputs("Astria supervision smoke daemon exited before health\n", stderr)
        }
        return 1
    }
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
            loadState.message = "Astria could not finish loading. Reload this window or open diagnostics."
        }

        func webView(_ webView: WKWebView, didFailProvisionalNavigation navigation: WKNavigation!, withError error: Error) {
            loadState.message = "Astria is waiting for the local daemon at \(AstriaDefaults.webURL)."
        }
    }
}
