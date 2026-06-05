#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const binDir = path.join(__dirname, '..', 'bin');
for (const name of ['starclaw-bin', 'starclaw.exe']) {
  const file = path.join(binDir, name);
  if (fs.existsSync(file)) {
    fs.rmSync(file, { force: true });
  }
}
console.log('StarClaw uninstalled');
