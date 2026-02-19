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

No programming tools needed — just download, unzip, and double-click.

<details>
<summary><strong>macOS (Apple Silicon — M1/M2/M3/M4)</strong></summary>

#### Step 1: Download the app

1. Go to the **[Releases page](../../releases/latest)**
2. Under **Assets**, click **`billing-checker-darwin-arm64`** to download it
3. Also click **Source code (zip)** to download the project files

#### Step 2: Set up the folder

1. Open **Finder** and go to your **Downloads** folder
2. Double-click the **Source code .zip** file to unzip it — this creates a folder called something like `anaesthetic-billing-checker-1.1.0`
3. **Rename** that folder to `Billing Checker` (or whatever you like) and move it somewhere convenient (e.g. your **Documents** folder)
4. Now **drag the `billing-checker-darwin-arm64` file** you downloaded into that folder (put it right next to the `app` folder and the `start-billing-app.command` file)
5. **Rename** the file from `billing-checker-darwin-arm64` to just `billing-checker` (remove the `-darwin-arm64` part)

Your folder should now look like this:
```
Billing Checker/
  ├── app/                          ← don't touch this
  ├── billing-checker               ← the file you downloaded and renamed
  ├── start-billing-app.command     ← double-click this to start
  ├── stop-billing-app.command      ← double-click this to stop
  └── ... (other files)
```

#### Step 3: Allow the app past macOS security

macOS blocks apps downloaded from the internet by default. You need to do this **once**:

1. Open **Terminal** (press ⌘+Space, type `Terminal`, press Enter)
2. Type the following command, then press Enter:
   ```
   xattr -cr ~/Documents/Billing\ Checker
   ```
   (If you put the folder somewhere else, adjust the path accordingly)
3. You can close Terminal — you won't need it again

> **What does this do?** It removes the "downloaded from the internet" quarantine flag so macOS will let the app run. This is safe — the app runs entirely on your computer and makes no network connections except to fetch official MBS data from the government website.

<details>
<summary><em>Alternative: if you don't want to use Terminal</em></summary>

1. In Finder, **right-click** (or Control-click) the file `billing-checker` and choose **Open**
2. macOS will show a warning dialog — click **Open** to allow it
3. A Terminal window will flash briefly and close — that's fine
4. Now **right-click** `start-billing-app.command` and choose **Open**
5. Again click **Open** in the warning dialog
6. From now on, you can just double-click to start

You may need to do this process twice if macOS asks again on the first real launch.
</details>

#### Step 4: Start the app

1. **Double-click** `start-billing-app.command`
2. A Terminal window will open showing "Anaesthetic Billing Checker" — **leave this window open** (you can minimise it)
3. Your browser will automatically open to the app at `http://localhost:8080`
4. On its very first launch, the app will download the latest MBS data — this takes about 30 seconds

#### Step 5: Stop the app

When you're done, either:
- **Double-click** `stop-billing-app.command`, or
- Close the Terminal window that opened in Step 4

</details>

<details>
<summary><strong>macOS (Intel)</strong></summary>

Follow the same steps as **macOS (Apple Silicon)** above, but in **Step 1** download **`billing-checker-darwin-amd64`** instead, and rename it to `billing-checker` in **Step 2**.

> **Not sure which Mac you have?** Click the Apple menu () → **About This Mac**. If it says "Apple M1" or "Apple M2" etc., use the Apple Silicon version. If it says "Intel", use this one.

</details>

<details>
<summary><strong>Windows</strong></summary>

#### Step 1: Download the app

1. Go to the **[Releases page](../../releases/latest)**
2. Under **Assets**, click **`billing-checker-windows-amd64.exe`** to download it
3. Also click **Source code (zip)** to download the project files

#### Step 2: Set up the folder

1. Open **File Explorer** and go to your **Downloads** folder
2. Right-click the **Source code .zip** file → **Extract All** → click **Extract**
3. **Rename** the extracted folder to `Billing Checker` and move it somewhere convenient (e.g. your **Documents** folder)
4. **Move the `billing-checker-windows-amd64.exe` file** into that folder (next to `start-billing-app.bat`)
5. **Rename** it from `billing-checker-windows-amd64.exe` to `billing-checker.exe`

Your folder should now look like this:
```
Billing Checker\
  ├── app\                          ← don't touch this
  ├── billing-checker.exe           ← the file you downloaded and renamed
  ├── start-billing-app.bat         ← double-click this to start
  ├── stop-billing-app.bat          ← double-click this to stop
  └── ... (other files)
```

#### Step 3: Allow the app past Windows security

When you first run the app, Windows SmartScreen may show a warning:

1. Click **More info**
2. Click **Run anyway**
3. You only need to do this once

#### Step 4: Start the app

1. **Double-click** `start-billing-app.bat`
2. A command window will open — **leave it open** (you can minimise it)
3. Your browser will automatically open to the app
4. On first launch, the app downloads the latest MBS data (~30 seconds)

#### Step 5: Stop the app

When you're done, either:
- **Double-click** `stop-billing-app.bat`, or
- Close the command window from Step 4

</details>

<details>
<summary><strong>Linux</strong></summary>

1. Download `billing-checker-linux-amd64` and `Source code (zip)` from the [Releases page](../../releases/latest)
2. Unzip the source code and place the binary inside the folder
3. Rename the binary to `billing-checker` and make it executable: `chmod +x billing-checker`
4. Run `./start-billing-app.sh`
5. Stop with `./stop-billing-app.sh` or `Ctrl+C`

</details>

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
