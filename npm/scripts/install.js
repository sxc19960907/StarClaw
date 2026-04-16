#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const binDir = path.join(__dirname, '..', 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

console.log('StarClaw binary would be downloaded here');
console.log('For development, build from source: go build .');
