#!/usr/bin/env node

const fs = require("node:fs");
const fsp = require("node:fs/promises");
const https = require("node:https");
const path = require("node:path");
const extractZip = require("extract-zip");
const tar = require("tar");

const pkg = require("../package.json");

const isWindows = process.platform === "win32";
const binaryName = isWindows ? "pong-terminal.exe" : "pong-terminal";
const vendorDir = path.join(__dirname, "..", "vendor");
const archMap = {
  x64: "amd64",
  arm64: "arm64"
};
const platformMap = {
  win32: "windows",
  linux: "linux",
  darwin: "darwin"
};

async function main() {
  const os = platformMap[process.platform];
  const arch = archMap[process.arch];

  if (!os) {
    throw new Error(`Unsupported OS: ${process.platform}`);
  }

  if (!arch) {
    throw new Error(`Unsupported architecture: ${process.arch}`);
  }

  if (os === "windows" && arch === "arm64") {
    throw new Error("Windows arm64 release is not available.");
  }

  const versionTag = `v${pkg.version}`;
  const ext = os === "windows" ? "zip" : "tar.gz";
  const asset = `pong-terminal_${versionTag}_${os}_${arch}.${ext}`;
  const url = `https://github.com/It-Shu/pong/releases/download/${versionTag}/${asset}`;

  await fsp.rm(vendorDir, { recursive: true, force: true });
  await fsp.mkdir(vendorDir, { recursive: true });

  const archivePath = path.join(vendorDir, asset);
  process.stdout.write(`Downloading ${asset}...\n`);
  await download(url, archivePath);

  if (os === "windows") {
    await extractZip(archivePath, { dir: vendorDir });
  } else {
    await tar.x({
      file: archivePath,
      cwd: vendorDir
    });
  }

  await fsp.rm(archivePath, { force: true });

  const installedBinary = path.join(vendorDir, binaryName);
  if (!fs.existsSync(installedBinary)) {
    throw new Error(`Installed binary not found: ${installedBinary}`);
  }

  if (!isWindows) {
    await fsp.chmod(installedBinary, 0o755);
  }
}

function download(url, destination) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        return resolve(download(response.headers.location, destination));
      }

      if (response.statusCode !== 200) {
        response.resume();
        return reject(new Error(`Download failed with status ${response.statusCode}`));
      }

      const file = fs.createWriteStream(destination);
      response.pipe(file);

      file.on("finish", () => {
        file.close(resolve);
      });

      file.on("error", (error) => {
        file.close(() => reject(error));
      });
    });

    request.on("error", reject);
  });
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
