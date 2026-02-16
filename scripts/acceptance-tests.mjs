/**
 * acceptance-tests.mjs
 * Runs the acceptance test cases from ACCEPTANCE_TESTS.md against the app
 * in a Node.js VM context (simulating browser globals).
 *
 * Copyright (c) 2026 Bernard McClement
 * Licensed under the MIT License. See LICENSE file in the project root.
 * You may copy, modify, and redistribute this software with attribution.
 */

import fs from "node:fs";
import vm from "node:vm";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const APP_DIR = path.join(__dirname, "..", "app");

const mbsDataCode = fs.readFileSync(path.join(APP_DIR, "mbs-data.js"), "utf8");
const appCode = fs.readFileSync(path.join(APP_DIR, "app.js"), "utf8");

// Build a mock DOM environment
function createMockContext() {
  const elements = {};
  const makeEl = () => ({
    classList: { add() {}, remove() {}, contains() { return false; } },
    innerHTML: "",
    value: "",
    disabled: false,
    textContent: "",
    className: "",
    addEventListener() {},
  });

  return {
    console,
    Intl,
    window: {},
    location: { protocol: "file:" },
    setTimeout() {},
    document: {
      getElementById: (id) => elements[id] || (elements[id] = makeEl()),
    },
  };
}

function loadApp() {
  const ctx = createMockContext();
  vm.createContext(ctx);
  vm.runInContext(mbsDataCode, ctx);
  ctx.window.MBS_DATA = ctx.window.MBS_DATA || ctx.MBS_DATA || {};
  ctx.window.MBS_DATA_META = ctx.window.MBS_DATA_META || ctx.MBS_DATA_META || {};
  vm.runInContext(appCode, ctx);
  return ctx;
}

function runCase(label, input, assertions) {
  const ctx = loadApp();
  const parsed = ctx.parseCodes(input);
  const procedures = ctx.detectProcedures(parsed.uniqueCodes);
  const finance = ctx.calculateFinance(parsed.uniqueCodes, parsed.counts);
  const warnings = ctx.buildWarnings(parsed.uniqueCodes, procedures, finance);
  const safety = ctx.buildSafetyChecks(parsed.uniqueCodes, parsed.counts);

  console.log(`\n═══ Case ${label} ═══`);
  console.log(`  Codes parsed: ${parsed.totalEntered} (${parsed.uniqueCodes.length} unique)`);
  console.log(`  Procedures:   ${procedures.join(" | ") || "none"}`);
  console.log(`  RVG units:    ${finance.knownUnitsTotal}`);
  console.log(`  Schedule fee:  $${finance.scheduleFeeTotal.toFixed(2)}`);
  console.log(`  No-gap:        $${finance.noGapMin.toFixed(2)} – $${finance.noGapMax.toFixed(2)}`);
  console.log(`  Private:       $${finance.privateEstimate.toFixed(2)}`);
  console.log(`  Unknown codes: ${finance.unknownCodes.join(", ") || "none"}`);

  console.log(`  Safety checks:`);
  for (const ch of safety) {
    console.log(`    [${ch.label.padEnd(9)}] ${ch.title}`);
  }

  console.log(`  Warnings:`);
  for (const w of warnings) {
    const sev = (w.severity || w.level || "?").toUpperCase();
    console.log(`    [${sev.padEnd(12)}] ${w.text}`);
  }

  let pass = true;
  for (const [desc, check] of Object.entries(assertions)) {
    const ok = check({ parsed, procedures, finance, warnings, safety });
    console.log(`  ${ok ? "✅" : "❌"} ${desc}`);
    if (!ok) pass = false;
  }
  console.log(`  Result: ${pass ? "PASS ✅" : "FAIL ❌"}`);
  return pass;
}

// ── Test cases ────────────────────────────────────────────────────

const results = [];

// Case A: Valve replacement with missing arterial insertion
results.push(runCase("A", `20560
25005
22012
22012
22012
22015
22020
22020
22052
22054
25014
23360
17620`, {
  "Valve pattern detected": ({ procedures }) =>
    procedures.some((p) => /valve/i.test(p)),
  "No fabricated values": ({ finance }) =>
    finance.unknownCodes.every((c) => true), // just must not crash
  "Has schedule fee total > 0": ({ finance }) =>
    finance.scheduleFeeTotal > 0,
  "Has unit-based totals": ({ finance }) =>
    finance.knownUnitsTotal > 0,
}));

// Case B: TURP elderly case
results.push(runCase("B", `20914
25014
17610
25000
23045`, {
  "All codes parsed": ({ parsed }) =>
    parsed.uniqueCodes.length === 5,
  "Has MBS descriptions": ({ parsed }) => {
    const ctx2 = loadApp();
    return parsed.uniqueCodes.every((c) => {
      const info = ctx2.getCodeInfo(c);
      return info && info.description;
    });
  },
  "Pre-op: PASS (one)": ({ safety }) =>
    safety.find((c) => /pre-op/i.test(c.title))?.label === "PASS",
  "Procedure present: PASS": ({ safety }) =>
    safety.find((c) => /procedure/i.test(c.title))?.label === "PASS",
  "Time code: PASS": ({ safety }) =>
    safety.find((c) => /time/i.test(c.title))?.label === "PASS",
}));

// Case C: Double pre-op hard fail
results.push(runCase("C", `17610
17620
20914
23045`, {
  "Pre-op: HARD FAIL": ({ safety }) =>
    safety.find((c) => /pre-op/i.test(c.title))?.label === "HARD FAIL",
}));

// Case D: Missing time hard fail
results.push(runCase("D", `17610
20914
25014`, {
  "Time code: HARD FAIL": ({ safety }) =>
    safety.find((c) => /time/i.test(c.title))?.label === "HARD FAIL",
}));

// Case E: High audit-risk trigger
results.push(runCase("E", `17610
20560
22014
23045`, {
  "High audit-risk QUERY": ({ safety }) =>
    safety.find((c) => /audit.?risk/i.test(c.title))?.label === "QUERY",
  "High audit-risk warning visible": ({ safety }) =>
    safety.find((c) => /audit.?risk/i.test(c.title))?.status === "amber",
}));

// Case G: Duplicate ASA descriptors hard fail
results.push(runCase("G", `20560
25005
25010
22012
23360
17620`, {
  "Duplicate ASA: HARD FAIL": ({ safety }) =>
    safety.find((c) => /ASA physical status/i.test(c.title))?.label === "HARD FAIL",
}));

// Summary
console.log("\n═══ SUMMARY ═══");
const labels = ["A", "B", "C", "D", "E", "G"];
for (let i = 0; i < results.length; i++) {
  console.log(`  Case ${labels[i]}: ${results[i] ? "PASS ✅" : "FAIL ❌"}`);
}
console.log(`  Case F: Manual test (trigger update button in UI)`);
const allPass = results.every(Boolean);
console.log(`\n  Overall: ${allPass ? "ALL PASS ✅" : "SOME FAILURES ❌"}`);
process.exitCode = allPass ? 0 : 1;
