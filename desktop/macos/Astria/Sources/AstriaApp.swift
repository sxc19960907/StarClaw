import SwiftUI
import WebKit
import Darwin

private enum AstriaDefaults {
    static let mainWindowID = "astria-main"
    static let webURL = "http://127.0.0.1:7533/app/"
    static let healthURL = "http://127.0.0.1:7533/health"
    static let diagnosticsURL = "http://127.0.0.1:7533/diagnostics"
    static let bundleIdentifier = "dev.starclaw.astria"
    static let desktopRPCProtocolVersion = "1.0.0"
    static let desktopRPCSocketName = "daemon.sock"
    static let desktopRPCPidfileName = "daemon.pid"
    static let startupTimeoutSeconds: TimeInterval = 5
    static let healthProbeTimeoutSeconds: TimeInterval = 0.5
    static let healthPollIntervalSeconds: TimeInterval = 0.12
    static let desktopRPCPollIntervalSeconds: TimeInterval = 2.0
    static let desktopRPCRetryDelaySeconds: TimeInterval = 0.25
    static let desktopRPCRetryLimit = 3
}

struct AstriaNativeCommand: Equatable {
    let id: String
    let title: String
    let key: String
    let modifiers: EventModifiers
}

enum AstriaNativeCommandSpec {
    static let newWindow = AstriaNativeCommand(id: "new_window", title: "New Window", key: "n", modifiers: [.command])
    static let reload = AstriaNativeCommand(id: "reload", title: "Reload Astria", key: "r", modifiers: [.command])
    static let diagnostics = AstriaNativeCommand(id: "diagnostics", title: "Open Diagnostics", key: "d", modifiers: [.command, .shift])
    static let exportDiagnostics = AstriaNativeCommand(id: "export_diagnostics", title: "Export Diagnostics", key: "e", modifiers: [.command, .shift])
    static let retryDaemon = AstriaNativeCommand(id: "retry_daemon", title: "Retry Daemon", key: "r", modifiers: [.command, .shift])

    static let all = [newWindow, reload, diagnostics, exportDiagnostics, retryDaemon]
}

@MainActor
final class AstriaAppActions: ObservableObject {
    var reload: () -> Void = {}
    var openDiagnostics: () -> Void = {}
    var exportDiagnostics: () -> Void = {}
    var retryDaemon: () -> Void = {}
}

struct AstriaNativeCommands: Commands {
    @Environment(\.openWindow) private var openWindow
    @EnvironmentObject private var appActions: AstriaAppActions

    var body: some Commands {
        CommandGroup(replacing: .newItem) {
            Button(AstriaNativeCommandSpec.newWindow.title) {
                openWindow(id: AstriaDefaults.mainWindowID)
            }
            .keyboardShortcut(KeyEquivalent(Character(AstriaNativeCommandSpec.newWindow.key)), modifiers: AstriaNativeCommandSpec.newWindow.modifiers)
        }

        CommandMenu("Astria") {
            Button(AstriaNativeCommandSpec.reload.title) {
                appActions.reload()
            }
            .keyboardShortcut(KeyEquivalent(Character(AstriaNativeCommandSpec.reload.key)), modifiers: AstriaNativeCommandSpec.reload.modifiers)

            Button(AstriaNativeCommandSpec.diagnostics.title) {
                appActions.openDiagnostics()
            }
            .keyboardShortcut(KeyEquivalent(Character(AstriaNativeCommandSpec.diagnostics.key)), modifiers: AstriaNativeCommandSpec.diagnostics.modifiers)

            Button(AstriaNativeCommandSpec.exportDiagnostics.title) {
                appActions.exportDiagnostics()
            }
            .keyboardShortcut(KeyEquivalent(Character(AstriaNativeCommandSpec.exportDiagnostics.key)), modifiers: AstriaNativeCommandSpec.exportDiagnostics.modifiers)

            Divider()

            Button(AstriaNativeCommandSpec.retryDaemon.title) {
                appActions.retryDaemon()
            }
            .keyboardShortcut(KeyEquivalent(Character(AstriaNativeCommandSpec.retryDaemon.key)), modifiers: AstriaNativeCommandSpec.retryDaemon.modifiers)
        }
    }
}

@main
struct AstriaApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var appActions = AstriaAppActions()

    init() {
        if CommandLine.arguments.contains("--native-command-smoke") {
            Foundation.exit(AstriaNativeCommandSmoke.run())
        }
        if CommandLine.arguments.contains("--diagnostics-export-smoke") {
            Foundation.exit(AstriaDiagnosticsExportSmoke.run())
        }
        if CommandLine.arguments.contains("--route-recovery-smoke") {
            Foundation.exit(AstriaRouteRecoverySmoke.run())
        }
        if CommandLine.arguments.contains("--desktop-rpc-smoke") {
            Foundation.exit(AstriaDesktopRPCSmoke.run())
        }
        if CommandLine.arguments.contains("--desktop-rpc-fallback-smoke") {
            Foundation.exit(AstriaDesktopRPCFallbackSmoke.run())
        }
        if CommandLine.arguments.contains("--desktop-rpc-session-smoke") {
            Foundation.exit(AstriaDesktopRPCSessionSmoke.run())
        }
        if CommandLine.arguments.contains("--supervision-smoke") {
            Foundation.exit(AstriaSupervisorSmoke.run())
        }
    }

    var body: some Scene {
        WindowGroup("Astria", id: AstriaDefaults.mainWindowID) {
            AstriaRootView(config: LaunchConfig.fromProcess())
                .environmentObject(appActions)
                .frame(minWidth: 1040, minHeight: 720)
        }
        .commands {
            AstriaNativeCommands()
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
    let runtimeDirectory: String
    let desktopRPCSocketPath: String
    let desktopRPCPidfilePath: String
    let appVersion: String
    let startupTimeout: TimeInterval
    let healthProbeTimeout: TimeInterval
    let healthPollInterval: TimeInterval
    let desktopRPCPollInterval: TimeInterval
    let desktopRPCRetryDelay: TimeInterval
    let desktopRPCRetryLimit: Int

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
        let runtimeDir = firstValue(after: "--runtime-dir", in: arguments)
            ?? environment["ASTRIA_RUNTIME_DIR"]
            ?? defaultRuntimeDirectory()
        let startupTimeout = timeIntervalValue(
            firstValue(after: "--startup-timeout", in: arguments)
                ?? environment["ASTRIA_STARTUP_TIMEOUT"],
            defaultValue: AstriaDefaults.startupTimeoutSeconds
        )
        let rpcPollInterval = timeIntervalValue(
            firstValue(after: "--desktop-rpc-poll-interval", in: arguments)
                ?? environment["ASTRIA_DESKTOP_RPC_POLL_INTERVAL"],
            defaultValue: AstriaDefaults.desktopRPCPollIntervalSeconds
        )
        let rpcRetryDelay = timeIntervalValue(
            firstValue(after: "--desktop-rpc-retry-delay", in: arguments)
                ?? environment["ASTRIA_DESKTOP_RPC_RETRY_DELAY"],
            defaultValue: AstriaDefaults.desktopRPCRetryDelaySeconds
        )
        let rpcRetryLimit = intValue(
            firstValue(after: "--desktop-rpc-retry-limit", in: arguments)
                ?? environment["ASTRIA_DESKTOP_RPC_RETRY_LIMIT"],
            defaultValue: AstriaDefaults.desktopRPCRetryLimit
        )

        return LaunchConfig(
            webURL: URL(string: webURLString) ?? URL(string: AstriaDefaults.webURL)!,
            healthURL: URL(string: healthURLString) ?? URL(string: AstriaDefaults.healthURL)!,
            diagnosticsURL: URL(string: diagnosticsURLString) ?? URL(string: AstriaDefaults.diagnosticsURL)!,
            starclawBinary: firstValue(after: "--starclaw-bin", in: arguments)
                ?? environment["ASTRIA_STARCLAW_BIN"],
            runtimeDirectory: runtimeDir,
            desktopRPCSocketPath: (runtimeDir as NSString).appendingPathComponent(AstriaDefaults.desktopRPCSocketName),
            desktopRPCPidfilePath: (runtimeDir as NSString).appendingPathComponent(AstriaDefaults.desktopRPCPidfileName),
            appVersion: bundleVersion(),
            startupTimeout: startupTimeout,
            healthProbeTimeout: AstriaDefaults.healthProbeTimeoutSeconds,
            healthPollInterval: AstriaDefaults.healthPollIntervalSeconds,
            desktopRPCPollInterval: rpcPollInterval,
            desktopRPCRetryDelay: rpcRetryDelay,
            desktopRPCRetryLimit: rpcRetryLimit
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

    private static func intValue(_ value: String?, defaultValue: Int) -> Int {
        guard let value, let parsed = Int(value), parsed > 0 else {
            return defaultValue
        }
        return parsed
    }

    private static func defaultRuntimeDirectory() -> String {
        let manager = FileManager.default
        if let support = manager.urls(for: .applicationSupportDirectory, in: .userDomainMask).first {
            return support.appendingPathComponent(AstriaDefaults.bundleIdentifier, isDirectory: true).path
        }
        return (NSTemporaryDirectory() as NSString).appendingPathComponent(AstriaDefaults.bundleIdentifier)
    }

    private static func bundleVersion() -> String {
        Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "0.0.0"
    }
}

struct AstriaRootView: View {
    let config: LaunchConfig
    @State private var loadState = WebLoadState()
    @State private var webURL: URL
    @State private var reloadToken = UUID()
    @StateObject private var supervisor: DaemonSupervisor
    @EnvironmentObject private var appActions: AstriaAppActions

    private let routeStore: AstriaRouteStore

    init(config: LaunchConfig) {
        let routeStore = AstriaRouteStore()
        self.config = config
        self.routeStore = routeStore
        _webURL = State(initialValue: routeStore.restoredURL(baseURL: config.webURL))
        _supervisor = StateObject(wrappedValue: DaemonSupervisor(config: config))
    }

    var body: some View {
        ZStack(alignment: .top) {
            if supervisor.state.isAttached {
                AstriaWebView(url: webURL, reloadToken: reloadToken, loadState: $loadState) { navigatedURL in
                    if AstriaRouteStore.route(from: navigatedURL, baseURL: config.webURL) != nil {
                        routeStore.persist(url: navigatedURL, baseURL: config.webURL)
                        webURL = navigatedURL
                    }
                }
                    .ignoresSafeArea()
            } else {
                LaunchStateView(state: supervisor.state, diagnosticsURL: config.diagnosticsURL) {
                    supervisor.start()
                }
            }
            if let message = loadState.message ?? supervisor.state.bannerMessage ?? DesktopRPCDiagnostics.bannerMessage(for: supervisor.desktopRPCSessionState) {
                StatusBanner(message: message, diagnosticsURL: config.diagnosticsURL, canRetry: supervisor.state.canRetry || DesktopRPCDiagnostics.canRetry(supervisor.desktopRPCSessionState)) {
                    supervisor.start()
                }
                    .padding(.top, 12)
                    .padding(.horizontal, 16)
            }
        }
        .task {
            registerNativeActions()
            supervisor.start()
        }
        .onDisappear {
            clearNativeActions()
        }
        .onChange(of: supervisor.recoveryGeneration) { _ in
            webURL = routeStore.restoredURL(baseURL: config.webURL)
            reloadToken = UUID()
        }
    }

    private func registerNativeActions() {
        appActions.reload = {
            reloadToken = UUID()
        }
        appActions.openDiagnostics = {
            NSWorkspace.shared.open(config.diagnosticsURL)
        }
        appActions.exportDiagnostics = {
            do {
                let url = try AstriaDiagnosticsExporter.export(
                    config: config,
                    daemonState: supervisor.state,
                    desktopRPCState: supervisor.desktopRPCSessionState
                )
                NSWorkspace.shared.activateFileViewerSelecting([url])
            } catch {
                loadState.message = "Astria could not export diagnostics. \(error.localizedDescription)"
            }
        }
        appActions.retryDaemon = {
            supervisor.start()
        }
    }

    private func clearNativeActions() {
        appActions.reload = {}
        appActions.openDiagnostics = {}
        appActions.exportDiagnostics = {}
        appActions.retryDaemon = {}
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
    case degraded(String)
    case unavailable
    case recovered
    case failed(String)
    case crashed(String)

    var isAttached: Bool {
        switch self {
        case .attached, .degraded, .unavailable, .recovered:
            return true
        default:
            return false
        }
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
        case .degraded:
            return "Opening Astria with HTTP fallback"
        case .unavailable:
            return "Reconnecting to StarClaw daemon"
        case .recovered:
            return "Restoring Astria"
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
        case .degraded(let message):
            return message
        case .unavailable:
            return "The local daemon is temporarily unavailable. Astria will reload when it recovers."
        case .recovered:
            return "The daemon recovered. Astria is reloading the last workspace route."
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
        case .degraded(let message):
            return message
        case .unavailable:
            return "Local daemon connection lost. Astria will recover when health returns."
        case .recovered:
            return "Local daemon recovered. Reloading Astria."
        case .failed(let message), .crashed(let message):
            return message
        default:
            return nil
        }
    }
}

enum DesktopRPCSessionState: Equatable {
    case idle
    case connecting
    case connected
    case reconnecting(Int)
    case degraded(String)
    case mismatch(String)
}

enum DesktopRPCDiagnostics {
    static func bannerMessage(for state: DesktopRPCSessionState) -> String? {
        switch state {
        case .idle, .connected:
            return nil
        case .connecting:
            return "Astria is validating the local Desktop RPC session."
        case .reconnecting(let attempt):
            return "Desktop RPC disconnected. Astria is retrying local session recovery (attempt \(attempt))."
        case .degraded(let message):
            return message
        case .mismatch(let message):
            return "Desktop RPC compatibility mismatch. Astria is using HTTP fallback. \(message)"
        }
    }

    static func severity(for state: DesktopRPCSessionState) -> String {
        switch state {
        case .idle:
            return "idle"
        case .connecting, .reconnecting:
            return "warning"
        case .connected:
            return "ready"
        case .degraded:
            return "degraded"
        case .mismatch:
            return "mismatch"
        }
    }

    static func canRetry(_ state: DesktopRPCSessionState) -> Bool {
        switch state {
        case .degraded, .mismatch:
            return true
        default:
            return false
        }
    }
}

struct AstriaDiagnosticsReport: Encodable {
    let generatedAt: String
    let appVersion: String
    let webURL: String
    let healthURL: String
    let diagnosticsURL: String
    let daemonState: String
    let desktopRPCState: String
    let failureSummary: String?
    let localOnly: Bool
}

enum AstriaDiagnosticsRedactor {
    private static let sensitiveKeys = [
        "api_key",
        "openai_api_key",
        "authorization",
        "bearer",
        "token",
        "secret",
        "password",
    ]

    static func redact(_ value: String, config: LaunchConfig? = nil) -> String {
        var redacted = value
        if let config {
            redacted = redacted.replacingOccurrences(of: config.desktopRPCSocketPath, with: "[redacted-desktop-rpc-socket]")
            redacted = redacted.replacingOccurrences(of: config.desktopRPCPidfilePath, with: "[redacted-desktop-rpc-pidfile]")
        }
        redacted = redactBearerTokens(in: redacted)
        for key in sensitiveKeys {
            redacted = redactAssignments(in: redacted, key: key)
        }
        return redacted
    }

    private static func redactAssignments(in value: String, key: String) -> String {
        let pattern = "(?i)(\(NSRegularExpression.escapedPattern(for: key))\\s*[:=]\\s*)([^\\s,;]+)"
        guard let regex = try? NSRegularExpression(pattern: pattern) else {
            return value
        }
        let range = NSRange(value.startIndex..<value.endIndex, in: value)
        return regex.stringByReplacingMatches(in: value, range: range, withTemplate: "$1[redacted]")
    }

    private static func redactBearerTokens(in value: String) -> String {
        let pattern = "(?i)(bearer\\s+)([A-Za-z0-9._~+\\-/=]+)"
        guard let regex = try? NSRegularExpression(pattern: pattern) else {
            return value
        }
        let range = NSRange(value.startIndex..<value.endIndex, in: value)
        return regex.stringByReplacingMatches(in: value, range: range, withTemplate: "$1[redacted]")
    }
}

enum AstriaDiagnosticsExporter {
    static func report(config: LaunchConfig, daemonState: DaemonState, desktopRPCState: DesktopRPCSessionState) -> AstriaDiagnosticsReport {
        let failureSummary = rawFailureSummary(daemonState: daemonState, desktopRPCState: desktopRPCState)
        return AstriaDiagnosticsReport(
            generatedAt: ISO8601DateFormatter().string(from: Date()),
            appVersion: config.appVersion,
            webURL: config.webURL.absoluteString,
            healthURL: config.healthURL.absoluteString,
            diagnosticsURL: config.diagnosticsURL.absoluteString,
            daemonState: label(for: daemonState),
            desktopRPCState: label(for: desktopRPCState),
            failureSummary: failureSummary.map { AstriaDiagnosticsRedactor.redact($0, config: config) },
            localOnly: true
        )
    }

    static func export(config: LaunchConfig, daemonState: DaemonState, desktopRPCState: DesktopRPCSessionState) throws -> URL {
        let directory = try diagnosticsDirectory(config: config)
        let filename = "astria-diagnostics-\(Int(Date().timeIntervalSince1970)).json"
        let url = directory.appendingPathComponent(filename)
        let data = try JSONEncoder.sortedPrettyPrinted.encode(report(config: config, daemonState: daemonState, desktopRPCState: desktopRPCState))
        try data.write(to: url, options: [.atomic])
        return url
    }

    static func diagnosticsDirectory(config: LaunchConfig) throws -> URL {
        let directory = URL(fileURLWithPath: config.runtimeDirectory, isDirectory: true)
            .appendingPathComponent("diagnostics", isDirectory: true)
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        return directory
    }

    private static func rawFailureSummary(daemonState: DaemonState, desktopRPCState: DesktopRPCSessionState) -> String? {
        switch daemonState {
        case .failed(let message), .crashed(let message), .degraded(let message):
            return message
        default:
            return DesktopRPCDiagnostics.bannerMessage(for: desktopRPCState)
        }
    }

    private static func label(for state: DaemonState) -> String {
        switch state {
        case .idle:
            return "idle"
        case .checking:
            return "checking"
        case .starting:
            return "starting"
        case .attached:
            return "attached"
        case .degraded:
            return "degraded"
        case .unavailable:
            return "unavailable"
        case .recovered:
            return "recovered"
        case .failed:
            return "failed"
        case .crashed:
            return "crashed"
        }
    }

    private static func label(for state: DesktopRPCSessionState) -> String {
        DesktopRPCDiagnostics.severity(for: state)
    }
}

extension JSONEncoder {
    static var sortedPrettyPrinted: JSONEncoder {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        return encoder
    }
}

@MainActor
final class DaemonSupervisor: ObservableObject {
    @Published private(set) var state: DaemonState = .idle
    @Published private(set) var desktopRPCSessionState: DesktopRPCSessionState = .idle
    @Published private(set) var recoveryGeneration = 0

    private let config: LaunchConfig
    private var childProcess: Process?
    private var launchTask: Task<Void, Never>?
    private var healthMonitorTask: Task<Void, Never>?
    private var desktopRPCMonitorTask: Task<Void, Never>?

    init(config: LaunchConfig) {
        self.config = config
    }

    func start() {
        launchTask?.cancel()
        healthMonitorTask?.cancel()
        stopDesktopRPCMonitor()
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
            if FileManager.default.fileExists(atPath: config.desktopRPCSocketPath) {
                do {
                    try await reconcileDesktopRPC()
                } catch {
                    state = .degraded("Desktop RPC reconciliation failed; Astria is using the healthy daemon over HTTP. \(error.localizedDescription)")
                    startHealthMonitor()
                    return
                }
            } else {
                state = .degraded("Desktop RPC is not available on the healthy daemon. Astria is using HTTP fallback.")
                startHealthMonitor()
                return
            }
            state = .attached
            startDesktopRPCMonitor()
            startHealthMonitor()
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
            do {
                try await reconcileDesktopRPC()
            } catch {
                state = .failed("Desktop RPC reconciliation failed: \(error.localizedDescription)")
                return
            }
            state = .attached
            startDesktopRPCMonitor()
            startHealthMonitor()
            return
        }

        let message = process.isRunning
            ? "Daemon did not become healthy within \(Int(config.startupTimeout)) seconds. Check whether port 7533 is already in use or open diagnostics."
            : "Daemon process exited before becoming healthy. Open diagnostics or run `starclaw app --check` in Terminal."
        state = .failed(message)
    }

    private func launchDaemon() throws -> Process {
        try AstriaRuntimeArtifacts(config: config).cleanStaleArtifacts()
        return try DaemonLaunchSupport.launchDaemon(config: config) { [weak self] proc in
            Task { @MainActor [weak self] in
                self?.handleDaemonExit(status: proc.terminationStatus)
            }
        }
    }

    private func handleDaemonExit(status: Int32) {
        childProcess = nil
        stopDesktopRPCMonitor()
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

    private func reconcileDesktopRPC() async throws {
        desktopRPCSessionState = .connecting
        let socketPath = config.desktopRPCSocketPath
        let appVersion = config.appVersion
        try await Task.detached(priority: .userInitiated) {
            let capabilities = try DesktopRPCClient(socketPath: socketPath).systemCapabilities()
            try DesktopRPCReconciler.validate(capabilities: capabilities, appVersion: appVersion)
        }.value
        desktopRPCSessionState = .connected
    }

    private func startDesktopRPCMonitor() {
        desktopRPCMonitorTask?.cancel()
        desktopRPCMonitorTask = Task { [weak self] in
            guard let self else {
                return
            }
            await self.monitorDesktopRPC()
        }
    }

    private func stopDesktopRPCMonitor() {
        desktopRPCMonitorTask?.cancel()
        desktopRPCMonitorTask = nil
        desktopRPCSessionState = .idle
    }

    private func monitorDesktopRPC() async {
        let pollNanoseconds = UInt64(max(config.desktopRPCPollInterval, 0.1) * 1_000_000_000)
        while !Task.isCancelled {
            try? await Task.sleep(nanoseconds: pollNanoseconds)
            if Task.isCancelled {
                return
            }
            let result = await DesktopRPCSessionProbe.probe(
                socketPath: config.desktopRPCSocketPath,
                appVersion: config.appVersion
            )
            switch result {
            case .connected:
                desktopRPCSessionState = .connected
                if case .degraded = state, await isHealthy() {
                    state = .attached
                }
            case .mismatch(let message):
                desktopRPCSessionState = .mismatch(message)
                state = .degraded("Desktop RPC compatibility mismatch; Astria is using HTTP fallback. \(message)")
                return
            case .recoverableFailure(let message):
                let recovered = await retryDesktopRPC(after: message)
                if !recovered {
                    return
                }
            }
        }
    }

    private func retryDesktopRPC(after message: String) async -> Bool {
        let result = await DesktopRPCSessionRecovery.recover(
            socketPath: config.desktopRPCSocketPath,
            appVersion: config.appVersion,
            retryLimit: config.desktopRPCRetryLimit,
            retryDelay: config.desktopRPCRetryDelay
        ) { attempt in
            await MainActor.run {
                self.desktopRPCSessionState = .reconnecting(attempt)
            }
        }
        switch result {
        case .connected:
            desktopRPCSessionState = .connected
            if case .degraded = state, await isHealthy() {
                state = .attached
            }
            return true
        case .mismatch(let message):
            desktopRPCSessionState = .mismatch(message)
            state = .degraded("Desktop RPC compatibility mismatch; Astria is using HTTP fallback. \(message)")
            return false
        case .recoverableFailure(let recoveredMessage):
            let degradedMessage = "Desktop RPC disconnected; Astria is using the healthy daemon over HTTP fallback. \(message); \(recoveredMessage)"
            desktopRPCSessionState = .degraded(degradedMessage)
            if await isHealthy() {
                state = .degraded(degradedMessage)
            }
            return false
        }
    }

    private func startHealthMonitor() {
        healthMonitorTask?.cancel()
        healthMonitorTask = Task { [weak self] in
            guard let self else {
                return
            }
            await self.monitorHealth()
        }
    }

    private func monitorHealth() async {
        var sawUnavailable = false
        while !Task.isCancelled {
            try? await Task.sleep(nanoseconds: UInt64(max(config.healthPollInterval * 6, 0.5) * 1_000_000_000))
            let healthy = await isHealthy()
            if healthy {
                if sawUnavailable {
                    sawUnavailable = false
                    state = .recovered
                    recoveryGeneration += 1
                    try? await Task.sleep(nanoseconds: 1_000_000_000)
                    if state == .recovered {
                        state = .attached
                    }
                }
                continue
            }

            sawUnavailable = true
            switch state {
            case .attached, .degraded, .recovered:
                state = .unavailable
            default:
                break
            }
        }
    }
}

struct AstriaRuntimeArtifacts {
    let runtimeDirectory: String
    let socketPath: String
    let pidfilePath: String

    init(config: LaunchConfig) {
        runtimeDirectory = config.runtimeDirectory
        socketPath = config.desktopRPCSocketPath
        pidfilePath = config.desktopRPCPidfilePath
    }

    init(runtimeDirectory: String) {
        self.runtimeDirectory = runtimeDirectory
        socketPath = (runtimeDirectory as NSString).appendingPathComponent(AstriaDefaults.desktopRPCSocketName)
        pidfilePath = (runtimeDirectory as NSString).appendingPathComponent(AstriaDefaults.desktopRPCPidfileName)
    }

    func cleanStaleArtifacts() throws {
        try FileManager.default.createDirectory(atPath: runtimeDirectory, withIntermediateDirectories: true)

        switch pidfileState() {
        case .live:
            return
        case .dead, .malformed:
            try removeOwnedArtifact(pidfilePath)
            try removeOwnedArtifact(socketPath)
        case .missing:
            break
        }

        if FileManager.default.fileExists(atPath: socketPath) {
            do {
                let capabilities = try DesktopRPCClient(socketPath: socketPath).systemCapabilities()
                try DesktopRPCReconciler.validate(capabilities: capabilities, appVersion: "0.0.0")
            } catch {
                try removeOwnedArtifact(socketPath)
            }
        }
    }

    func isOwnedArtifact(_ path: String) -> Bool {
        let standardizedRuntime = (runtimeDirectory as NSString).standardizingPath
        let standardizedPath = (path as NSString).standardizingPath
        return standardizedPath == (standardizedRuntime as NSString).appendingPathComponent(AstriaDefaults.desktopRPCSocketName)
            || standardizedPath == (standardizedRuntime as NSString).appendingPathComponent(AstriaDefaults.desktopRPCPidfileName)
    }

    func removeOwnedArtifact(_ path: String) throws {
        guard isOwnedArtifact(path) else {
            throw DesktopRPCError.unsafeArtifactPath(path)
        }
        do {
            try FileManager.default.removeItem(atPath: path)
        } catch CocoaError.fileNoSuchFile {
        } catch {
            throw error
        }
    }

    private func pidfileState() -> PidfileState {
        guard let contents = try? String(contentsOfFile: pidfilePath) else {
            return .missing
        }
        let trimmed = contents.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let pid = Int32(trimmed), pid > 0 else {
            return .malformed
        }
        if kill(pid, 0) == 0 || errno == EPERM {
            return .live
        }
        return .dead
    }

    private enum PidfileState {
        case missing
        case live
        case dead
        case malformed
    }
}

enum DaemonLaunchSupport {
    static func launchDaemon(config: LaunchConfig, terminationHandler: ((Process) -> Void)? = nil) throws -> Process {
        try FileManager.default.createDirectory(
            atPath: (config.desktopRPCSocketPath as NSString).deletingLastPathComponent,
            withIntermediateDirectories: true
        )
        let process = Process()
        let resolved = resolveStarclawBinary(config: config)
        process.executableURL = resolved.executableURL
        process.arguments = resolved.arguments + [
            "daemon", "start",
            "--rpc-socket", config.desktopRPCSocketPath,
            "--rpc-pidfile", config.desktopRPCPidfilePath,
        ]

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

struct DesktopRPCCapabilities {
    let version: String
    let methods: [String]
    let appVersion: String
}

enum DesktopRPCReconciler {
    static func validate(capabilities: DesktopRPCCapabilities, appVersion: String) throws {
        guard capabilities.version == AstriaDefaults.desktopRPCProtocolVersion else {
            throw DesktopRPCError.protocolMismatch(expected: AstriaDefaults.desktopRPCProtocolVersion, got: capabilities.version)
        }

        let methods = Set(capabilities.methods)
        for method in ["system.ping", "system.capabilities"] where !methods.contains(method) {
            throw DesktopRPCError.missingMethod(method)
        }

        if isReleaseVersion(appVersion), isReleaseVersion(capabilities.appVersion), appVersion != capabilities.appVersion {
            throw DesktopRPCError.versionMismatch(app: appVersion, daemon: capabilities.appVersion)
        }
    }

    private static func isReleaseVersion(_ version: String) -> Bool {
        let trimmed = version.trimmingCharacters(in: .whitespacesAndNewlines)
        return !trimmed.isEmpty && trimmed != "dev" && trimmed != "0" && trimmed != "0.0.0"
    }
}

enum DesktopRPCError: LocalizedError {
    case socketPathTooLong(String)
    case connectFailed(String)
    case writeFailed
    case readFailed
    case invalidFrame
    case rpcError(String)
    case protocolMismatch(expected: String, got: String)
    case missingMethod(String)
    case versionMismatch(app: String, daemon: String)
    case unsafeArtifactPath(String)

    var errorDescription: String? {
        switch self {
        case .socketPathTooLong(let path):
            return "Desktop RPC socket path is too long: \(path)"
        case .connectFailed(let path):
            return "Could not connect to Desktop RPC socket at \(path)."
        case .writeFailed:
            return "Could not write Desktop RPC request."
        case .readFailed:
            return "Could not read Desktop RPC response."
        case .invalidFrame:
            return "Desktop RPC returned an invalid frame."
        case .rpcError(let message):
            return "Desktop RPC returned an error: \(message)"
        case .protocolMismatch(let expected, let got):
            return "Desktop RPC protocol mismatch: expected \(expected), got \(got)."
        case .missingMethod(let method):
            return "Desktop RPC missing required method \(method)."
        case .versionMismatch(let app, let daemon):
            return "Astria app version \(app) is not compatible with daemon version \(daemon)."
        case .unsafeArtifactPath(let path):
            return "Refusing to remove Desktop RPC artifact outside Astria runtime ownership: \(path)"
        }
    }
}

enum DesktopRPCSessionProbeResult: Equatable {
    case connected
    case recoverableFailure(String)
    case mismatch(String)
}

enum DesktopRPCSessionProbe {
    static func probe(socketPath: String, appVersion: String) async -> DesktopRPCSessionProbeResult {
        await Task.detached(priority: .utility) {
            do {
                let capabilities = try DesktopRPCClient(socketPath: socketPath).systemCapabilities()
                try DesktopRPCReconciler.validate(capabilities: capabilities, appVersion: appVersion)
                return .connected
            } catch let error as DesktopRPCError {
                switch error {
                case .protocolMismatch, .missingMethod, .versionMismatch:
                    return .mismatch(error.localizedDescription)
                default:
                    return .recoverableFailure(error.localizedDescription)
                }
            } catch {
                return .recoverableFailure(error.localizedDescription)
            }
        }.value
    }
}

enum DesktopRPCSessionRecovery {
    static func recover(
        socketPath: String,
        appVersion: String,
        retryLimit: Int,
        retryDelay: TimeInterval,
        onAttempt: @escaping (Int) async -> Void = { _ in }
    ) async -> DesktopRPCSessionProbeResult {
        let attempts = max(retryLimit, 1)
        let delay = UInt64(max(retryDelay, 0.05) * 1_000_000_000)
        var lastResult: DesktopRPCSessionProbeResult = .recoverableFailure("Desktop RPC reconnect did not run")
        for attempt in 1...attempts {
            await onAttempt(attempt)
            try? await Task.sleep(nanoseconds: delay)
            if Task.isCancelled {
                return .recoverableFailure("Desktop RPC reconnect cancelled")
            }
            let result = await DesktopRPCSessionProbe.probe(socketPath: socketPath, appVersion: appVersion)
            switch result {
            case .connected, .mismatch:
                return result
            case .recoverableFailure:
                lastResult = result
            }
        }
        return lastResult
    }
}

struct DesktopRPCClient {
    let socketPath: String

    func systemCapabilities() throws -> DesktopRPCCapabilities {
        let fd = try connect()
        defer {
            close(fd)
        }

        let requestID = "astria_caps_\(UUID().uuidString)"
        let payload: [String: Any] = [
            "request_id": requestID,
            "method": "system.capabilities",
            "source": "astria",
        ]
        let frame: [String: Any] = [
            "type": "desktop_rpc_request",
            "payload": payload,
        ]
        try writeFrame(frame, to: fd)
        let response = try readFrame(from: fd)

        guard
            let type = response["type"] as? String,
            type == "desktop_rpc_result",
            let resultPayload = response["payload"] as? [String: Any],
            resultPayload["request_id"] as? String == requestID
        else {
            throw DesktopRPCError.invalidFrame
        }
        if let ok = resultPayload["ok"] as? Bool, !ok {
            let error = resultPayload["error"] as? [String: Any]
            throw DesktopRPCError.rpcError(error?["message"] as? String ?? "unknown")
        }
        guard
            let result = resultPayload["result"] as? [String: Any],
            let version = result["version"] as? String,
            let methods = result["methods"] as? [String],
            let platform = result["platform"] as? [String: Any]
        else {
            throw DesktopRPCError.invalidFrame
        }

        return DesktopRPCCapabilities(
            version: version,
            methods: methods,
            appVersion: platform["app_version"] as? String ?? "dev"
        )
    }

    private func connect() throws -> Int32 {
        let maxPathLength = MemoryLayout.size(ofValue: sockaddr_un().sun_path)
        guard socketPath.utf8.count < maxPathLength else {
            throw DesktopRPCError.socketPathTooLong(socketPath)
        }

        let fd = socket(AF_UNIX, SOCK_STREAM, 0)
        guard fd >= 0 else {
            throw DesktopRPCError.connectFailed(socketPath)
        }

        var address = sockaddr_un()
        address.sun_family = sa_family_t(AF_UNIX)
        let pathBytes = Array(socketPath.utf8) + [0]
        withUnsafeMutableBytes(of: &address.sun_path) { rawBuffer in
            rawBuffer.copyBytes(from: pathBytes)
        }

        let length = socklen_t(MemoryLayout<sa_family_t>.size + pathBytes.count)
        let result = withUnsafePointer(to: &address) { pointer in
            pointer.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                Darwin.connect(fd, $0, length)
            }
        }
        guard result == 0 else {
            close(fd)
            throw DesktopRPCError.connectFailed(socketPath)
        }
        return fd
    }

    private func writeFrame(_ frame: [String: Any], to fd: Int32) throws {
        let body = try JSONSerialization.data(withJSONObject: frame)
        var length = UInt32(body.count).bigEndian
        var data = Data(bytes: &length, count: 4)
        data.append(body)
        try data.withUnsafeBytes { rawBuffer in
            guard let base = rawBuffer.baseAddress else {
                throw DesktopRPCError.writeFailed
            }
            var sent = 0
            while sent < rawBuffer.count {
                let result = Darwin.write(fd, base.advanced(by: sent), rawBuffer.count - sent)
                guard result > 0 else {
                    throw DesktopRPCError.writeFailed
                }
                sent += result
            }
        }
    }

    private func readFrame(from fd: Int32) throws -> [String: Any] {
        let header = try readExact(4, from: fd)
        let length = header.reduce(UInt32(0)) { ($0 << 8) | UInt32($1) }
        guard length > 0, length <= 4 * 1024 * 1024 else {
            throw DesktopRPCError.invalidFrame
        }
        let body = try readExact(Int(length), from: fd)
        guard
            let object = try JSONSerialization.jsonObject(with: Data(body)) as? [String: Any]
        else {
            throw DesktopRPCError.invalidFrame
        }
        return object
    }

    private func readExact(_ count: Int, from fd: Int32) throws -> [UInt8] {
        var buffer = [UInt8](repeating: 0, count: count)
        var offset = 0
        while offset < count {
            let result = buffer.withUnsafeMutableBytes { rawBuffer in
                Darwin.read(fd, rawBuffer.baseAddress!.advanced(by: offset), count - offset)
            }
            guard result > 0 else {
                throw DesktopRPCError.readFailed
            }
            offset += result
        }
        return buffer
    }
}

struct AstriaRouteStore {
    static let key = "astria.lastWebRoute"

    private let defaults: UserDefaults

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    func persist(url: URL, baseURL: URL) {
        guard let route = Self.route(from: url, baseURL: baseURL) else {
            return
        }
        defaults.set(route, forKey: Self.key)
    }

    func restoredURL(baseURL: URL) -> URL {
        guard
            let route = defaults.string(forKey: Self.key),
            let url = Self.url(forRoute: route, baseURL: baseURL)
        else {
            return baseURL
        }
        return url
    }

    static func route(from url: URL, baseURL: URL) -> String? {
        guard sameOrigin(url, baseURL) else {
            return nil
        }
        return normalizedRoute(path: url.path, query: url.query, fragment: url.fragment)
    }

    static func url(forRoute route: String, baseURL: URL) -> URL? {
        guard route.hasPrefix("/") else {
            return nil
        }
        var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false)
        let split = splitRoute(route)
        guard let normalized = normalizedRoute(path: split.path, query: split.query, fragment: split.fragment) else {
            return nil
        }
        let normalizedSplit = splitRoute(normalized)
        components?.path = normalizedSplit.path
        components?.query = normalizedSplit.query
        components?.fragment = normalizedSplit.fragment
        return components?.url
    }

    private static func sameOrigin(_ lhs: URL, _ rhs: URL) -> Bool {
        lhs.scheme == rhs.scheme && lhs.host == rhs.host && lhs.port == rhs.port
    }

    private static func normalizedRoute(path: String, query: String?, fragment: String?) -> String? {
        let normalizedPath: String
        switch path {
        case "/app":
            normalizedPath = "/app/"
        default:
            guard path == "/app/" || path.hasPrefix("/app/") else {
                return nil
            }
            normalizedPath = path
        }

        var route = normalizedPath
        if let query, !query.isEmpty {
            route += "?\(query)"
        }
        if let fragment, !fragment.isEmpty {
            route += "#\(fragment)"
        }
        return route
    }

    private static func splitRoute(_ route: String) -> (path: String, query: String?, fragment: String?) {
        let parts = route.split(separator: "#", maxSplits: 1, omittingEmptySubsequences: false)
        let routeWithoutFragment = String(parts[0])
        let fragment = parts.count > 1 ? String(parts[1]) : nil
        let queryParts = routeWithoutFragment.split(separator: "?", maxSplits: 1, omittingEmptySubsequences: false)
        let path = String(queryParts[0])
        let query = queryParts.count > 1 ? String(queryParts[1]) : nil
        return (path: path, query: query, fragment: fragment)
    }
}

enum AstriaSupervisorSmoke {
    static func run() -> Int32 {
        let config = LaunchConfig.fromProcess()

        if DaemonLaunchSupport.isHealthy(url: config.healthURL, timeout: config.healthProbeTimeout) {
            if FileManager.default.fileExists(atPath: config.desktopRPCSocketPath) {
                return reconcile(config: config)
            }
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
            return reconcile(config: config)
        }

        if process.isRunning {
            fputs("Astria supervision smoke timed out waiting for daemon health\n", stderr)
        } else {
            fputs("Astria supervision smoke daemon exited before health\n", stderr)
        }
        return 1
    }

    private static func reconcile(config: LaunchConfig) -> Int32 {
        do {
            let capabilities = try DesktopRPCClient(socketPath: config.desktopRPCSocketPath).systemCapabilities()
            try DesktopRPCReconciler.validate(capabilities: capabilities, appVersion: config.appVersion)
            return 0
        } catch {
            fputs("Astria supervision smoke Desktop RPC reconciliation failed: \(error.localizedDescription)\n", stderr)
            return 1
        }
    }
}

enum AstriaNativeCommandSmoke {
    static func run() -> Int32 {
        let commands = AstriaNativeCommandSpec.all
        let ids = Set(commands.map(\.id))
        if commands.count != 5 || ids.count != commands.count {
            fputs("Astria native command smoke found missing or duplicate commands\n", stderr)
            return 1
        }
        let expected: [String: (title: String, key: String)] = [
            "new_window": ("New Window", "n"),
            "reload": ("Reload Astria", "r"),
            "diagnostics": ("Open Diagnostics", "d"),
            "export_diagnostics": ("Export Diagnostics", "e"),
            "retry_daemon": ("Retry Daemon", "r"),
        ]
        for command in commands {
            guard let want = expected[command.id] else {
                fputs("Astria native command smoke found unexpected command \(command.id)\n", stderr)
                return 1
            }
            if command.title != want.title || command.key != want.key {
                fputs("Astria native command smoke command \(command.id) = \(command.title)/\(command.key)\n", stderr)
                return 1
            }
            if command.modifiers.isEmpty {
                fputs("Astria native command smoke command \(command.id) has no shortcut modifiers\n", stderr)
                return 1
            }
        }
        if AstriaNativeCommandSpec.diagnostics.modifiers == AstriaNativeCommandSpec.reload.modifiers {
            fputs("Astria native command smoke diagnostics shortcut conflicts with reload\n", stderr)
            return 1
        }
        if AstriaNativeCommandSpec.retryDaemon.modifiers == AstriaNativeCommandSpec.reload.modifiers {
            fputs("Astria native command smoke retry shortcut conflicts with reload\n", stderr)
            return 1
        }
        return 0
    }
}

enum AstriaDiagnosticsExportSmoke {
    static func run() -> Int32 {
        let runtime = "/tmp/astria-diagnostics-\(getpid())-\(UUID().uuidString.prefix(8))"
        defer {
            try? FileManager.default.removeItem(atPath: runtime)
        }
        let config = LaunchConfig.fromProcess(arguments: [
            "Astria",
            "--runtime-dir", runtime,
        ], environment: [:])
        let sensitive = "failed with api_key=sk-test-secret Authorization: Bearer abc.def token=hidden \(config.desktopRPCSocketPath) \(config.desktopRPCPidfilePath)"
        let redacted = AstriaDiagnosticsRedactor.redact(sensitive, config: config)
        for forbidden in ["sk-test-secret", "abc.def", "hidden", config.desktopRPCSocketPath, config.desktopRPCPidfilePath] {
            if redacted.contains(forbidden) {
                fputs("Astria diagnostics export smoke redaction leaked \(forbidden)\n", stderr)
                return 1
            }
        }

        do {
            let url = try AstriaDiagnosticsExporter.export(
                config: config,
                daemonState: .failed(sensitive),
                desktopRPCState: .mismatch("token=hidden")
            )
            let body = try String(contentsOf: url)
            for required in ["\"localOnly\" : true", "\"daemonState\" : \"failed\"", "\"desktopRPCState\" : \"mismatch\""] {
                if !body.contains(required) {
                    fputs("Astria diagnostics export smoke missing \(required)\n", stderr)
                    return 1
                }
            }
            for forbidden in ["sk-test-secret", "abc.def", "hidden", config.desktopRPCSocketPath, config.desktopRPCPidfilePath] {
                if body.contains(forbidden) {
                    fputs("Astria diagnostics export smoke report leaked \(forbidden)\n", stderr)
                    return 1
                }
            }
        } catch {
            fputs("Astria diagnostics export smoke failed: \(error.localizedDescription)\n", stderr)
            return 1
        }
        return 0
    }
}

enum AstriaDesktopRPCSmoke {
    static func run() -> Int32 {
        let valid = DesktopRPCCapabilities(
            version: AstriaDefaults.desktopRPCProtocolVersion,
            methods: ["system.ping", "system.capabilities"],
            appVersion: "dev"
        )
        do {
            try DesktopRPCReconciler.validate(capabilities: valid, appVersion: "0.0.0")
        } catch {
            fputs("Astria Desktop RPC smoke rejected valid capabilities: \(error.localizedDescription)\n", stderr)
            return 1
        }

        if accepts(DesktopRPCCapabilities(version: "9.9.9", methods: valid.methods, appVersion: "dev"), appVersion: "0.0.0") {
            fputs("Astria Desktop RPC smoke accepted protocol mismatch\n", stderr)
            return 1
        }
        if accepts(DesktopRPCCapabilities(version: valid.version, methods: ["system.ping"], appVersion: "dev"), appVersion: "0.0.0") {
            fputs("Astria Desktop RPC smoke accepted missing capabilities method\n", stderr)
            return 1
        }
        if accepts(DesktopRPCCapabilities(version: valid.version, methods: valid.methods, appVersion: "1.0.0"), appVersion: "2.0.0") {
            fputs("Astria Desktop RPC smoke accepted release version mismatch\n", stderr)
            return 1
        }

        return 0
    }

    private static func accepts(_ capabilities: DesktopRPCCapabilities, appVersion: String) -> Bool {
        do {
            try DesktopRPCReconciler.validate(capabilities: capabilities, appVersion: appVersion)
            return true
        } catch {
            return false
        }
    }
}

enum AstriaDesktopRPCSessionSmoke {
    static func run() -> Int32 {
        if !validateDiagnostics() {
            return 1
        }

        let runtime = "/tmp/astria-session-\(getpid())-\(UUID().uuidString.prefix(8))"
        let socketPath = (runtime as NSString).appendingPathComponent(AstriaDefaults.desktopRPCSocketName)
        defer {
            try? FileManager.default.removeItem(atPath: runtime)
        }

        do {
            try FileManager.default.createDirectory(atPath: runtime, withIntermediateDirectories: true)
            try runCapabilitiesServer(socketPath: socketPath, protocolVersion: AstriaDefaults.desktopRPCProtocolVersion)
            let connected = awaitProbe(socketPath: socketPath, appVersion: "0.0.0")
            guard connected == .connected else {
                fputs("Astria Desktop RPC session smoke did not connect: \(connected)\n", stderr)
                return 1
            }

            let missing = awaitProbe(socketPath: socketPath, appVersion: "0.0.0")
            guard case .recoverableFailure = missing else {
                fputs("Astria Desktop RPC session smoke did not classify missing socket as recoverable: \(missing)\n", stderr)
                return 1
            }
            let recoveredMissing = awaitRecover(socketPath: socketPath, retryLimit: 2)
            guard case .recoverableFailure = recoveredMissing.result, recoveredMissing.attempts == [1, 2] else {
                fputs("Astria Desktop RPC session smoke did not bound reconnect attempts: \(recoveredMissing)\n", stderr)
                return 1
            }

            try runCapabilitiesServer(socketPath: socketPath, protocolVersion: "9.9.9")
            let mismatch = awaitProbe(socketPath: socketPath, appVersion: "0.0.0")
            guard case .mismatch = mismatch else {
                fputs("Astria Desktop RPC session smoke did not classify protocol mismatch: \(mismatch)\n", stderr)
                return 1
            }
        } catch {
            fputs("Astria Desktop RPC session smoke failed: \(error.localizedDescription)\n", stderr)
            return 1
        }

        return 0
    }

    private static func validateDiagnostics() -> Bool {
        if DesktopRPCDiagnostics.bannerMessage(for: .connected) != nil || DesktopRPCDiagnostics.severity(for: .connected) != "ready" {
            fputs("Astria Desktop RPC session smoke produced connected warning\n", stderr)
            return false
        }
        let reconnecting = DesktopRPCDiagnostics.bannerMessage(for: .reconnecting(2)) ?? ""
        if !reconnecting.contains("retrying local session recovery") || DesktopRPCDiagnostics.severity(for: .reconnecting(2)) != "warning" {
            fputs("Astria Desktop RPC session smoke missing reconnecting diagnostic\n", stderr)
            return false
        }
        let degraded = DesktopRPCDiagnostics.bannerMessage(for: .degraded("Desktop RPC disconnected; Astria is using HTTP fallback.")) ?? ""
        if !degraded.contains("HTTP fallback") || !DesktopRPCDiagnostics.canRetry(.degraded("x")) {
            fputs("Astria Desktop RPC session smoke missing degraded diagnostic\n", stderr)
            return false
        }
        let mismatch = DesktopRPCDiagnostics.bannerMessage(for: .mismatch("expected 1.0.0")) ?? ""
        if !mismatch.contains("compatibility mismatch") || DesktopRPCDiagnostics.severity(for: .mismatch("x")) != "mismatch" || !DesktopRPCDiagnostics.canRetry(.mismatch("x")) {
            fputs("Astria Desktop RPC session smoke missing mismatch diagnostic\n", stderr)
            return false
        }
        return true
    }

    private static func awaitProbe(socketPath: String, appVersion: String) -> DesktopRPCSessionProbeResult {
        let semaphore = DispatchSemaphore(value: 0)
        final class Box {
            var result: DesktopRPCSessionProbeResult = .recoverableFailure("probe did not complete")
        }
        let box = Box()
        Task {
            box.result = await DesktopRPCSessionProbe.probe(socketPath: socketPath, appVersion: appVersion)
            semaphore.signal()
        }
        if semaphore.wait(timeout: .now() + 5) == .timedOut {
            return .recoverableFailure("probe timed out")
        }
        return box.result
    }

    private static func awaitRecover(socketPath: String, retryLimit: Int) -> (result: DesktopRPCSessionProbeResult, attempts: [Int]) {
        let semaphore = DispatchSemaphore(value: 0)
        final class Box {
            var result: DesktopRPCSessionProbeResult = .recoverableFailure("recovery did not complete")
            var attempts: [Int] = []
        }
        let box = Box()
        Task {
            box.result = await DesktopRPCSessionRecovery.recover(
                socketPath: socketPath,
                appVersion: "0.0.0",
                retryLimit: retryLimit,
                retryDelay: 0.05
            ) { attempt in
                box.attempts.append(attempt)
            }
            semaphore.signal()
        }
        if semaphore.wait(timeout: .now() + 5) == .timedOut {
            return (.recoverableFailure("recovery timed out"), box.attempts)
        }
        return (box.result, box.attempts)
    }

    private static func runCapabilitiesServer(socketPath: String, protocolVersion: String) throws {
        let maxPathLength = MemoryLayout.size(ofValue: sockaddr_un().sun_path)
        guard socketPath.utf8.count < maxPathLength else {
            throw DesktopRPCError.socketPathTooLong(socketPath)
        }
        try? FileManager.default.removeItem(atPath: socketPath)
        let fd = socket(AF_UNIX, SOCK_STREAM, 0)
        guard fd >= 0 else {
            throw DesktopRPCError.connectFailed(socketPath)
        }
        var address = sockaddr_un()
        address.sun_family = sa_family_t(AF_UNIX)
        let pathBytes = Array(socketPath.utf8) + [0]
        withUnsafeMutableBytes(of: &address.sun_path) { rawBuffer in
            rawBuffer.copyBytes(from: pathBytes)
        }
        let length = socklen_t(MemoryLayout<sa_family_t>.size + pathBytes.count)
        let bindResult = withUnsafePointer(to: &address) { pointer in
            pointer.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                Darwin.bind(fd, $0, length)
            }
        }
        guard bindResult == 0 else {
            close(fd)
            throw DesktopRPCError.connectFailed(socketPath)
        }
        guard listen(fd, 1) == 0 else {
            close(fd)
            throw DesktopRPCError.connectFailed(socketPath)
        }

        DispatchQueue.global(qos: .userInitiated).async {
            let conn = accept(fd, nil, nil)
            defer {
                close(fd)
                try? FileManager.default.removeItem(atPath: socketPath)
            }
            guard conn >= 0 else {
                return
            }
            defer {
                close(conn)
            }
            guard let request = try? readFrame(from: conn),
                  let payload = request["payload"] as? [String: Any],
                  let requestID = payload["request_id"] as? String
            else {
                return
            }
            let response: [String: Any] = [
                "type": "desktop_rpc_result",
                "payload": [
                    "request_id": requestID,
                    "ok": true,
                    "result": [
                        "version": protocolVersion,
                        "methods": ["system.ping", "system.capabilities"],
                        "platform": [
                            "os": "macOS",
                            "os_version": "",
                            "app_version": "dev",
                        ],
                    ],
                ],
            ]
            try? writeFrame(response, to: conn)
        }
    }

    private static func readFrame(from fd: Int32) throws -> [String: Any] {
        let header = try readExact(4, from: fd)
        let length = header.reduce(UInt32(0)) { ($0 << 8) | UInt32($1) }
        let body = try readExact(Int(length), from: fd)
        guard let object = try JSONSerialization.jsonObject(with: Data(body)) as? [String: Any] else {
            throw DesktopRPCError.invalidFrame
        }
        return object
    }

    private static func readExact(_ count: Int, from fd: Int32) throws -> [UInt8] {
        var buffer = [UInt8](repeating: 0, count: count)
        var offset = 0
        while offset < count {
            let result = buffer.withUnsafeMutableBytes { rawBuffer in
                Darwin.read(fd, rawBuffer.baseAddress!.advanced(by: offset), count - offset)
            }
            guard result > 0 else {
                throw DesktopRPCError.readFailed
            }
            offset += result
        }
        return buffer
    }

    private static func writeFrame(_ frame: [String: Any], to fd: Int32) throws {
        let body = try JSONSerialization.data(withJSONObject: frame)
        var length = UInt32(body.count).bigEndian
        var data = Data(bytes: &length, count: 4)
        data.append(body)
        try data.withUnsafeBytes { rawBuffer in
            guard let base = rawBuffer.baseAddress else {
                throw DesktopRPCError.writeFailed
            }
            var sent = 0
            while sent < rawBuffer.count {
                let result = Darwin.write(fd, base.advanced(by: sent), rawBuffer.count - sent)
                guard result > 0 else {
                    throw DesktopRPCError.writeFailed
                }
                sent += result
            }
        }
    }
}

enum AstriaDesktopRPCFallbackSmoke {
    static func run() -> Int32 {
        let manager = FileManager.default
        let runtime = (NSTemporaryDirectory() as NSString).appendingPathComponent("astria-fallback-smoke-\(UUID().uuidString)")
        let artifacts = AstriaRuntimeArtifacts(runtimeDirectory: runtime)
        defer {
            try? manager.removeItem(atPath: runtime)
        }

        do {
            try manager.createDirectory(atPath: runtime, withIntermediateDirectories: true)
            try "999999\n".write(toFile: artifacts.pidfilePath, atomically: true, encoding: .utf8)
            try "stale".write(toFile: artifacts.socketPath, atomically: true, encoding: .utf8)
            try artifacts.cleanStaleArtifacts()
            if manager.fileExists(atPath: artifacts.pidfilePath) || manager.fileExists(atPath: artifacts.socketPath) {
                fputs("Astria Desktop RPC fallback smoke did not remove dead pidfile artifacts\n", stderr)
                return 1
            }

            try manager.createDirectory(atPath: runtime, withIntermediateDirectories: true)
            try "\(getpid())\n".write(toFile: artifacts.pidfilePath, atomically: true, encoding: .utf8)
            try "live".write(toFile: artifacts.socketPath, atomically: true, encoding: .utf8)
            try artifacts.cleanStaleArtifacts()
            if !manager.fileExists(atPath: artifacts.pidfilePath) || !manager.fileExists(atPath: artifacts.socketPath) {
                fputs("Astria Desktop RPC fallback smoke removed live pid artifacts\n", stderr)
                return 1
            }
            try artifacts.removeOwnedArtifact(artifacts.pidfilePath)
            try artifacts.removeOwnedArtifact(artifacts.socketPath)

            try "stale".write(toFile: artifacts.socketPath, atomically: true, encoding: .utf8)
            try artifacts.cleanStaleArtifacts()
            if manager.fileExists(atPath: artifacts.socketPath) {
                fputs("Astria Desktop RPC fallback smoke did not remove stale socket under runtime\n", stderr)
                return 1
            }

            let outside = (NSTemporaryDirectory() as NSString).appendingPathComponent("astria-outside-\(UUID().uuidString)")
            try "outside".write(toFile: outside, atomically: true, encoding: .utf8)
            defer {
                try? manager.removeItem(atPath: outside)
            }
            do {
                try artifacts.removeOwnedArtifact(outside)
                fputs("Astria Desktop RPC fallback smoke removed unsafe path\n", stderr)
                return 1
            } catch DesktopRPCError.unsafeArtifactPath {
                if !manager.fileExists(atPath: outside) {
                    fputs("Astria Desktop RPC fallback smoke deleted unsafe path despite error\n", stderr)
                    return 1
                }
            }
        } catch {
            fputs("Astria Desktop RPC fallback smoke failed: \(error.localizedDescription)\n", stderr)
            return 1
        }

        return 0
    }
}

enum AstriaRouteRecoverySmoke {
    static func run() -> Int32 {
        guard let baseURL = URL(string: AstriaDefaults.webURL) else {
            fputs("Astria route smoke could not parse base URL\n", stderr)
            return 1
        }
        let suiteName = "dev.starclaw.astria.route-smoke.\(UUID().uuidString)"
        guard let defaults = UserDefaults(suiteName: suiteName) else {
            fputs("Astria route smoke could not create defaults suite\n", stderr)
            return 1
        }
        defer {
            defaults.removePersistentDomain(forName: suiteName)
        }
        let store = AstriaRouteStore(defaults: defaults)

        let validURL = URL(string: "http://127.0.0.1:7533/app/?view=mission#runs")!
        store.persist(url: validURL, baseURL: baseURL)
        guard store.restoredURL(baseURL: baseURL).absoluteString == validURL.absoluteString else {
            fputs("Astria route smoke failed to restore valid route\n", stderr)
            return 1
        }

        let externalURL = URL(string: "https://example.com/app/#bad")!
        store.persist(url: externalURL, baseURL: baseURL)
        guard store.restoredURL(baseURL: baseURL).absoluteString == validURL.absoluteString else {
            fputs("Astria route smoke persisted external route\n", stderr)
            return 1
        }

        defaults.set("https://example.com/app/#bad", forKey: AstriaRouteStore.key)
        guard store.restoredURL(baseURL: baseURL).absoluteString == baseURL.absoluteString else {
            fputs("Astria route smoke did not fall back for unsafe stored route\n", stderr)
            return 1
        }

        defaults.set("/app#compact", forKey: AstriaRouteStore.key)
        guard store.restoredURL(baseURL: baseURL).absoluteString == "http://127.0.0.1:7533/app/#compact" else {
            fputs("Astria route smoke did not normalize /app route\n", stderr)
            return 1
        }

        return 0
    }
}

struct AstriaWebView: NSViewRepresentable {
    let url: URL
    let reloadToken: UUID
    @Binding var loadState: WebLoadState
    let onNavigationFinished: (URL) -> Void

    func makeCoordinator() -> Coordinator {
        Coordinator(loadState: $loadState, onNavigationFinished: onNavigationFinished, reloadToken: reloadToken)
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
        if context.coordinator.consumeReloadToken(reloadToken) {
            view.load(URLRequest(url: url))
            return
        }
        if view.url == nil {
            view.load(URLRequest(url: url))
        }
    }

    final class Coordinator: NSObject, WKNavigationDelegate {
        @Binding private var loadState: WebLoadState
        private let onNavigationFinished: (URL) -> Void
        private var reloadToken: UUID

        init(loadState: Binding<WebLoadState>, onNavigationFinished: @escaping (URL) -> Void, reloadToken: UUID) {
            _loadState = loadState
            self.onNavigationFinished = onNavigationFinished
            self.reloadToken = reloadToken
        }

        func consumeReloadToken(_ token: UUID) -> Bool {
            guard token != reloadToken else {
                return false
            }
            reloadToken = token
            return true
        }

        func webView(_ webView: WKWebView, didStartProvisionalNavigation navigation: WKNavigation!) {
            loadState.message = nil
        }

        func webView(_ webView: WKWebView, didFinish navigation: WKNavigation!) {
            loadState.message = nil
            if let url = webView.url {
                onNavigationFinished(url)
            }
        }

        func webView(_ webView: WKWebView, didFail navigation: WKNavigation!, withError error: Error) {
            loadState.message = "Astria could not finish loading. Reload this window or open diagnostics."
        }

        func webView(_ webView: WKWebView, didFailProvisionalNavigation navigation: WKNavigation!, withError error: Error) {
            loadState.message = "Astria is waiting for the local daemon at \(AstriaDefaults.webURL)."
        }
    }
}
