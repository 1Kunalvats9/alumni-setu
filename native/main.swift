import Cocoa
import WebKit

let apiPort = 8080
let appURL = URL(string: "http://localhost:\(apiPort)/")!
let healthURL = URL(string: "http://localhost:\(apiPort)/api/health")!

final class AppDelegate: NSObject, NSApplicationDelegate, WKNavigationDelegate {
    var window: NSWindow!
    var webView: WKWebView!
    var server: Process?

    func applicationDidFinishLaunching(_ notification: Notification) {
        buildMenu()
        buildWindow()

        DispatchQueue.global(qos: .userInitiated).async {
            if !self.serverIsUp() {
                self.startServer()
            }
            let ready = self.waitForServer(attempts: 40)

            DispatchQueue.main.async {
                if ready {
                    self.webView.load(URLRequest(url: appURL))
                } else {
                    self.showStartupFailure()
                }
            }
        }
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        return true
    }

    func applicationWillTerminate(_ notification: Notification) {
        stopServer()
    }

    private func buildWindow() {
        window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 1280, height: 860),
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        window.title = "AlumniSetu"
        window.minSize = NSSize(width: 900, height: 600)
        window.center()
        window.setFrameAutosaveName("AlumniSetuMainWindow")

        let configuration = WKWebViewConfiguration()
        configuration.websiteDataStore = .default()

        webView = WKWebView(frame: window.contentView!.bounds, configuration: configuration)
        webView.autoresizingMask = [.width, .height]
        webView.navigationDelegate = self
        webView.setValue(false, forKey: "drawsBackground")

        window.contentView!.addSubview(webView)
        window.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }

    private func buildMenu() {
        let mainMenu = NSMenu()

        let appMenuItem = NSMenuItem()
        let appMenu = NSMenu()
        appMenu.addItem(withTitle: "About AlumniSetu", action: #selector(NSApplication.orderFrontStandardAboutPanel(_:)), keyEquivalent: "")
        appMenu.addItem(.separator())
        appMenu.addItem(withTitle: "Hide AlumniSetu", action: #selector(NSApplication.hide(_:)), keyEquivalent: "h")
        appMenu.addItem(withTitle: "Quit AlumniSetu", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q")
        appMenuItem.submenu = appMenu
        mainMenu.addItem(appMenuItem)

        let editMenuItem = NSMenuItem()
        let editMenu = NSMenu(title: "Edit")
        editMenu.addItem(withTitle: "Undo", action: Selector(("undo:")), keyEquivalent: "z")
        editMenu.addItem(withTitle: "Redo", action: Selector(("redo:")), keyEquivalent: "Z")
        editMenu.addItem(.separator())
        editMenu.addItem(withTitle: "Cut", action: #selector(NSText.cut(_:)), keyEquivalent: "x")
        editMenu.addItem(withTitle: "Copy", action: #selector(NSText.copy(_:)), keyEquivalent: "c")
        editMenu.addItem(withTitle: "Paste", action: #selector(NSText.paste(_:)), keyEquivalent: "v")
        editMenu.addItem(withTitle: "Select All", action: #selector(NSText.selectAll(_:)), keyEquivalent: "a")
        editMenuItem.submenu = editMenu
        mainMenu.addItem(editMenuItem)

        let viewMenuItem = NSMenuItem()
        let viewMenu = NSMenu(title: "View")
        viewMenu.addItem(withTitle: "Reload", action: #selector(reload), keyEquivalent: "r")
        viewMenuItem.submenu = viewMenu
        mainMenu.addItem(viewMenuItem)

        NSApp.mainMenu = mainMenu
    }

    @objc private func reload() {
        webView.reload()
    }

    private func supportDirectory() -> URL {
        let base = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask)[0]
        let directory = base.appendingPathComponent("AlumniSetu", isDirectory: true)
        try? FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        return directory
    }

    private func startServer() {
        guard let resources = Bundle.main.resourcePath else { return }

        let binary = URL(fileURLWithPath: resources).appendingPathComponent("server/alumnisetu-server")
        guard FileManager.default.fileExists(atPath: binary.path) else {
            NSLog("no bundled server at %@", binary.path)
            return
        }

        let webRoot = URL(fileURLWithPath: resources).appendingPathComponent("renderer")
        let database = supportDirectory().appendingPathComponent("alumnisetu.db")

        var environment = ProcessInfo.processInfo.environment
        environment["ALUMNISETU_DB"] = database.path
        environment["ALUMNISETU_WEB"] = webRoot.path

        let process = Process()
        process.executableURL = binary
        process.environment = environment

        do {
            try process.run()
            server = process
            NSLog("server started, database at %@", database.path)
        } catch {
            NSLog("could not start the server: %@", error.localizedDescription)
        }
    }

    private func stopServer() {
        server?.terminate()
        server = nil
    }

    private func serverIsUp() -> Bool {
        var request = URLRequest(url: healthURL)
        request.timeoutInterval = 1

        let semaphore = DispatchSemaphore(value: 0)
        var reachable = false

        URLSession.shared.dataTask(with: request) { _, response, _ in
            if let http = response as? HTTPURLResponse, http.statusCode == 200 {
                reachable = true
            }
            semaphore.signal()
        }.resume()

        _ = semaphore.wait(timeout: .now() + 2)
        return reachable
    }

    private func waitForServer(attempts: Int) -> Bool {
        for _ in 0..<attempts {
            if serverIsUp() { return true }
            Thread.sleep(forTimeInterval: 0.25)
        }
        return false
    }

    private func showStartupFailure() {
        let alert = NSAlert()
        alert.messageText = "AlumniSetu cannot reach its server"
        alert.informativeText = "Nothing is answering on port \(apiPort).\n\n"
            + "If you are running from source, start it with:\n    cd server && go run ."
        alert.alertStyle = .critical
        alert.addButton(withTitle: "Quit")
        alert.runModal()
        NSApp.terminate(nil)
    }
}

let application = NSApplication.shared
let delegate = AppDelegate()
application.delegate = delegate
application.setActivationPolicy(.regular)
application.run()
