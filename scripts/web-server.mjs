/**
 * web-server.mjs
 * Local HTTP server for Anaesthetic Billing Checker.
 * Serves static frontend from app/ and provides API endpoints
 * for MBS catalog update and status.
 *
 * Copyright (c) 2026 Bernard McClement
 * Licensed under the MIT License. See LICENSE file in the project root.
 * You may copy, modify, and redistribute this software with attribution.
 */

import fs from "node:fs/promises";
import path from "node:path";
import { createServer } from "node:http";
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT_DIR = path.resolve(__dirname, "..");
const APP_DIR = path.join(ROOT_DIR, "app");
const MBS_DATA_FILE = path.join(APP_DIR, "mbs-data.js");

const PORT = Number(process.env.PORT || 8080);

let updateInProgress = false;
let firstOpenDone = false;

const MIME = {
  ".html": "text/html; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".txt": "text/plain; charset=utf-8",
};

// ── Server ────────────────────────────────────────────────────────
const server = createServer(async (req, res) => {
  const url = new URL(req.url || "/", `http://${req.headers.host}`);

  // API: catalog metadata
  if (req.method === "GET" && url.pathname === "/api/meta") {
    await triggerFirstOpenUpdate();
    const meta = await readMeta();
    return json(res, 200, { ok: true, meta, updateInProgress });
  }

  // API: trigger MBS catalog update
  if (req.method === "POST" && url.pathname === "/api/update-mbs") {
    if (updateInProgress) return json(res, 409, { ok: false, error: "Update already in progress." });
    try {
      await runMbsUpdate();
      const meta = await readMeta();
      return json(res, 200, { ok: true, meta });
    } catch (err) {
      return json(res, 500, { ok: false, error: err.message || "Update failed." });
    }
  }

  // Serve index.html
  if (req.method === "GET" && (url.pathname === "/" || url.pathname === "/index.html")) {
    triggerFirstOpenUpdate().catch(() => {});
    return serveFile(path.join(APP_DIR, "index.html"), res);
  }

  // Serve other static files
  if (req.method === "GET") {
    const safe = safePath(url.pathname);
    if (!safe) return text(res, 400, "Bad request");
    return serveFile(path.join(APP_DIR, safe), res);
  }

  text(res, 405, "Method not allowed");
});

server.listen(PORT, () => {
  console.log(`\n  Anaesthetic Billing Checker`);
  console.log(`  Running at  http://localhost:${PORT}`);
  console.log(`  First page open triggers automatic MBS data update.\n`);
});

// ── Auto-update on first open ─────────────────────────────────────
async function triggerFirstOpenUpdate() {
  if (firstOpenDone) return;
  firstOpenDone = true;
  if (process.env.DISABLE_AUTO_UPDATE === "1") return;
  try {
    await runMbsUpdate();
  } catch (err) {
    console.error("Auto-update failed:", err.message || err);
  }
}

// ── MBS update via child process ──────────────────────────────────
async function runMbsUpdate() {
  if (updateInProgress) throw new Error("Update already in progress.");
  updateInProgress = true;
  try {
    await exec("node", [path.join(ROOT_DIR, "scripts", "update-mbs-data.mjs")], ROOT_DIR);
  } finally {
    updateInProgress = false;
  }
}

function exec(cmd, args, cwd) {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args, { cwd, stdio: "pipe" });
    let stderr = "";
    child.stdout.on("data", (d) => process.stdout.write(d));
    child.stderr.on("data", (d) => {
      stderr += d.toString();
      process.stderr.write(d);
    });
    child.on("close", (code) =>
      code === 0 ? resolve() : reject(new Error(stderr.trim() || `Exit code ${code}`))
    );
  });
}

// ── Read catalog metadata from generated JS ───────────────────────
async function readMeta() {
  try {
    const content = await fs.readFile(MBS_DATA_FILE, "utf8");
    const m = content.match(/window\.MBS_DATA_META\s*=\s*(\{[\s\S]*?\});/);
    if (!m) return { source: "not-loaded", generatedAt: null, totalItems: 0 };
    return JSON.parse(m[1]);
  } catch {
    return { source: "not-loaded", generatedAt: null, totalItems: 0 };
  }
}

// ── Static file serving ───────────────────────────────────────────
function safePath(pathname) {
  const cleaned = pathname.replace(/^\/+/, "");
  if (!cleaned) return "index.html";
  if (cleaned.includes("..")) return null;
  return cleaned;
}

async function serveFile(filePath, res) {
  try {
    const data = await fs.readFile(filePath);
    const ext = path.extname(filePath).toLowerCase();
    res.writeHead(200, { "Content-Type": MIME[ext] || "application/octet-stream" });
    res.end(data);
  } catch {
    text(res, 404, "Not found");
  }
}

function json(res, status, payload) {
  res.writeHead(status, { "Content-Type": "application/json; charset=utf-8" });
  res.end(JSON.stringify(payload));
}

function text(res, status, body) {
  res.writeHead(status, { "Content-Type": "text/plain; charset=utf-8" });
  res.end(body);
}
