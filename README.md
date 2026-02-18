# Anaesthetic Billing Checker

A local tool for checking Australian MBS anaesthesia item numbers before claiming. Paste your codes, get an instant audit with safety checks, warnings, and fee estimates.

**No data leaves your computer.** Everything runs locally in your browser and a small local server.

<p align="center">
  <img src="docs/screenshot-start.png" alt="Paste your MBS codes" width="600" />
  <br/>
  <img src="docs/screenshot-result.png" alt="Audit result with safety checks and fee estimates" width="600" />
</p>

---

## What it does

- Parses MBS item numbers from pasted text (comma, space, or newline separated)
- Looks up official MBS descriptions and schedule fees
- Calculates MBS schedule fee total, no-gap estimate, and private estimate
- Runs traffic-light safety checks (PASS / QUERY / HARD FAIL)
- Flags audit-risk codes and billing consistency issues
- Shows a per-item code breakdown table
- Configurable RVG base rate and gap percentage

> **Important:** This is a decision-support tool only — not legal or billing advice. Always validate against the [official MBS Online](https://www.mbsonline.gov.au) and current AMA/ASA RVG references.

---

## Getting started

There are two ways to run the app — pick whichever suits you:

### Option A: Standalone binary (no install required)

Download a single pre-built binary — no Node.js, Python, or any runtime needed.

1. Go to the [Releases](../../releases) page
2. Download the binary for your platform:

   | Platform | Binary |
   |---|---|
   | macOS (Apple Silicon) | `billing-checker-darwin-arm64` |
   | macOS (Intel) | `billing-checker-darwin-amd64` |
   | Windows | `billing-checker-windows-amd64.exe` |
   | Linux | `billing-checker-linux-amd64` |

3. Place the binary in the project folder (alongside the `app/` directory) and rename it to `billing-checker` (or `billing-checker.exe` on Windows)
4. Double-click **start-billing-app.command** (macOS), **.bat** (Windows), or run `./start-billing-app.sh` (Linux)

> On first run, macOS may ask you to allow the app. Right-click → Open, then click Open in the dialog.

### Option B: Node.js (for development)

If you have **Node.js 18+** installed (or want to modify the server code):

<details>
<summary><strong>Installing Node.js</strong> (one-time, if not already installed)</summary>

**macOS — easiest:**
1. Go to [https://nodejs.org](https://nodejs.org) and click the green **LTS** button
2. Double-click the downloaded `.pkg` and follow the prompts

**macOS — Homebrew:** `brew install node`

**Windows:**
1. Go to [https://nodejs.org](https://nodejs.org) and click the green **LTS** button
2. Run the `.msi` installer with defaults

**Linux:**
```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
source ~/.bashrc
nvm install --lts
```
</details>

Then:
```bash
npm install        # first time only
npm start
```
Open [http://localhost:8080](http://localhost:8080). Press `Ctrl+C` to stop.

> **Tip:** The start scripts auto-detect which mode to use. If a `billing-checker` binary is present, they use it; otherwise they fall back to Node.js.

### Start & stop

| Action | macOS | Windows | Linux |
|---|---|---|---|
| Start | Double-click `start-billing-app.command` | Double-click `start-billing-app.bat` | `./start-billing-app.sh` |
| Stop | Double-click `stop-billing-app.command` | Double-click `stop-billing-app.bat` | `./stop-billing-app.sh` or `Ctrl+C` |

Use the `PORT` environment variable to change the port: `PORT=9090 ./billing-checker`

---

## Keeping MBS data up to date

The app automatically checks for the latest MBS data on first launch. You can also update manually:

- **In the app:** Click the **Update MBS Catalog** button
- **From terminal (Node.js mode):** `npm run update:mbs`

Data comes from the official [MBS Online XML downloads](https://www.mbsonline.gov.au/internet/mbsonline/publishing.nsf/Content/downloads).

---

## Project structure

```
├── app/
│   ├── index.html            Main page
│   ├── styles.css             Styling
│   ├── app.js                 All checker logic
│   └── mbs-data.js            Generated MBS data (auto-created on first run)
├── scripts/                   Node.js server & tools
│   ├── web-server.mjs         Local HTTP server
│   ├── update-mbs-data.mjs    Fetches official MBS XML
│   └── acceptance-tests.mjs   Automated test suite
├── main.go                    Standalone server + MBS updater (Go)
├── go.mod                     Go module definition
├── build.sh                   Cross-compilation script (Go)
├── package.json               Node.js project config
├── start-billing-app.*        One-click start scripts
├── stop-billing-app.*         One-click stop scripts
├── LICENSE
└── README.md
```

---

## Building standalone binaries from source

Requires [Go 1.21+](https://go.dev/dl/) (build-time only — end users don't need Go).

**Build for your current platform:**

```bash
go build -o billing-checker .
```

**Cross-compile for all platforms:**

```bash
./build.sh
```

This produces binaries in `dist/` for macOS (arm64 + amd64), Linux, and Windows.

---

## Running tests

```bash
node scripts/acceptance-tests.mjs
```

See [ACCEPTANCE_TESTS.md](ACCEPTANCE_TESTS.md) for the test case definitions.

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with clear, focused commits
4. Run tests: `node scripts/acceptance-tests.mjs`
5. Open a pull request

---

## License

[MIT](LICENSE) — see [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md) for dependency licenses.
