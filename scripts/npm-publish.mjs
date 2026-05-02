#!/usr/bin/env node
// Build + publish the revcat npm packages.
//
// Usage:
//   scripts/npm-publish.mjs <version> [--dry-run]
//
// Expects goreleaser to have produced archives in dist/:
//   dist/revcat_<version>_<os>_<arch>.tar.gz   (linux + darwin)
//   dist/revcat_<version>_windows_<arch>.zip   (windows)
//
// Walks the matrix in PLATFORMS, extracts each archive into the matching
// npm/<pkg>/bin/ dir, bumps every package.json to <version> (including
// the main package's optionalDependencies map so versions stay in lockstep),
// then `npm publish --access public` each platform package, and the main
// package last (so the optionalDependency entries already exist by the time
// someone installs revcat).
//
// --dry-run skips the publish step but still touches the files so you can
// inspect the result before pushing.

import { execSync } from "node:child_process";
import { readFileSync, writeFileSync, existsSync, mkdirSync, chmodSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const repo = resolve(__dirname, "..");
const distDir = resolve(repo, "dist");

const PLATFORMS = [
  { pkg: "revcat-darwin-arm64", goos: "darwin",  goarch: "arm64", binary: "revcat" },
  { pkg: "revcat-darwin-x64",   goos: "darwin",  goarch: "amd64", binary: "revcat" },
  { pkg: "revcat-linux-arm64",  goos: "linux",   goarch: "arm64", binary: "revcat" },
  { pkg: "revcat-linux-x64",    goos: "linux",   goarch: "amd64", binary: "revcat" },
  { pkg: "revcat-win32-x64",    goos: "windows", goarch: "amd64", binary: "revcat.exe" },
];

const args = process.argv.slice(2);
const version = args.find(a => !a.startsWith("--"));
const dryRun = args.includes("--dry-run");

if (!version || !/^\d+\.\d+\.\d+(?:-[\w.]+)?$/.test(version)) {
  console.error("usage: scripts/npm-publish.mjs <version> [--dry-run]");
  console.error("  version must be semver without the 'v' prefix (e.g. 0.6.0)");
  process.exit(1);
}

console.log(`> revcat npm publish - version=${version} dry_run=${dryRun}`);

// 1. extract each archive into npm/<pkg>/bin/
for (const p of PLATFORMS) {
  const ext = p.goos === "windows" ? "zip" : "tar.gz";
  const archive = resolve(distDir, `revcat_${version}_${p.goos}_${p.goarch}.${ext}`);
  const binDir = resolve(repo, "npm", p.pkg, "bin");
  if (!existsSync(archive)) {
    console.error(`missing archive: ${archive} - run goreleaser first`);
    process.exit(1);
  }
  mkdirSync(binDir, { recursive: true });
  if (ext === "zip") {
    execSync(`unzip -o -j "${archive}" "${p.binary}" -d "${binDir}"`, { stdio: "inherit" });
  } else {
    execSync(`tar -xzf "${archive}" -C "${binDir}" "${p.binary}"`, { stdio: "inherit" });
  }
  chmodSync(resolve(binDir, p.binary), 0o755);
  console.log(`  extracted ${p.pkg}/${p.binary}`);
}

// 2. bump versions in each package.json
const bump = (relPath, mut) => {
  const path = resolve(repo, relPath);
  const json = JSON.parse(readFileSync(path, "utf8"));
  mut(json);
  writeFileSync(path, JSON.stringify(json, null, 2) + "\n");
};
for (const p of PLATFORMS) {
  bump(`npm/${p.pkg}/package.json`, j => { j.version = version; });
}
bump("npm/revcat/package.json", j => {
  j.version = version;
  for (const p of PLATFORMS) j.optionalDependencies[p.pkg] = version;
});
console.log(`  bumped all package.json to ${version}`);

// 3. publish (platform packages first, main last)
const publish = (relPath) => {
  const cwd = resolve(repo, relPath);
  const cmd = "npm publish --access public" + (dryRun ? " --dry-run" : "");
  console.log(`> ${cmd}  (in ${relPath})`);
  execSync(cmd, { cwd, stdio: "inherit" });
};
for (const p of PLATFORMS) publish(`npm/${p.pkg}`);
publish("npm/revcat");

console.log("> done.");
