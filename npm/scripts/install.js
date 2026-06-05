#!/usr/bin/env node
const fs = require("fs");
const os = require("os");
const path = require("path");
const { execFileSync } = require("child_process");
const http = require("http");
const https = require("https");

const packageRoot = path.resolve(__dirname, "..");
const binDir = path.join(packageRoot, "bin");
const archiveOverride = process.env.STARCLAW_NPM_ARCHIVE || "";
const version = process.env.STARCLAW_NPM_VERSION || "latest";
const baseURL = process.env.STARCLAW_NPM_BASE_URL || "https://github.com/starclaw/starclaw/releases";

function platformAsset() {
  const platformMap = {
    darwin: "Darwin",
    linux: "Linux",
    win32: "Windows",
  };
  const archMap = {
    arm64: "arm64",
    x64: "x86_64",
  };
  const releaseOS = platformMap[process.platform];
  const releaseArch = archMap[process.arch];
  if (!releaseOS || !releaseArch) {
    throw new Error(`unsupported platform: ${process.platform}/${process.arch}`);
  }
  const ext = process.platform === "win32" ? "zip" : "tar.gz";
  return {
    archiveName: `starclaw_${releaseOS}_${releaseArch}.${ext}`,
    binaryName: process.platform === "win32" ? "starclaw.exe" : "starclaw",
    installedName: process.platform === "win32" ? "starclaw.exe" : "starclaw-bin",
  };
}

function releaseURL(archiveName) {
  if (version === "latest") {
    return `${baseURL}/latest/download/${archiveName}`;
  }
  return `${baseURL}/download/${version}/${archiveName}`;
}

function download(url, destination) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith("http://") ? http : https;
    const request = client.get(url, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        download(response.headers.location, destination).then(resolve, reject);
        return;
      }
      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`download ${url} returned ${response.statusCode}`));
        return;
      }
      const file = fs.createWriteStream(destination);
      response.pipe(file);
      file.on("finish", () => file.close(resolve));
      file.on("error", reject);
    });
    request.on("error", reject);
  });
}

function extractArchive(archivePath, destination) {
  fs.mkdirSync(destination, { recursive: true });
  if (archivePath.endsWith(".zip")) {
    if (process.platform === "win32") {
      execFileSync("powershell.exe", [
        "-NoProfile",
        "-Command",
        "Expand-Archive",
        "-LiteralPath",
        archivePath,
        "-DestinationPath",
        destination,
        "-Force",
      ]);
    } else {
      execFileSync("unzip", ["-q", archivePath, "-d", destination]);
    }
    return;
  }
  execFileSync("tar", ["-xzf", archivePath, "-C", destination]);
}

function findBinary(root, binaryName) {
  const entries = fs.readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    const candidate = path.join(root, entry.name);
    if (entry.isDirectory()) {
      const found = findBinary(candidate, binaryName);
      if (found) return found;
    }
    if (entry.isFile() && entry.name === binaryName) {
      return candidate;
    }
  }
  return "";
}

async function main() {
  const asset = platformAsset();
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "starclaw-npm-"));
  const archivePath = archiveOverride || path.join(tmpDir, asset.archiveName);
  const extractDir = path.join(tmpDir, "extract");
  try {
    if (!archiveOverride) {
      const url = releaseURL(asset.archiveName);
      console.log(`Downloading StarClaw ${version} from ${url}`);
      await download(url, archivePath);
    }
    extractArchive(archivePath, extractDir);
    const binary = findBinary(extractDir, asset.binaryName);
    if (!binary) {
      throw new Error(`archive does not contain ${asset.binaryName}`);
    }
    fs.mkdirSync(binDir, { recursive: true });
    const installedPath = path.join(binDir, asset.installedName);
    fs.copyFileSync(binary, installedPath);
    if (process.platform !== "win32") {
      fs.chmodSync(installedPath, 0o755);
    }
    console.log(`StarClaw installed to ${installedPath}`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

main().catch((error) => {
  console.error(`StarClaw npm install failed: ${error.message}`);
  console.error("Install from GitHub Releases or build from source instead:");
  console.error("  https://github.com/starclaw/starclaw/releases");
  console.error("  go install github.com/starclaw/starclaw@latest");
  process.exit(1);
});
