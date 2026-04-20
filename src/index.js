#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const os = require('os');
const { spawn } = require('child_process');
const omelette = require('omelette');

const { showHelp } = require('./lib/help');
const { setupCompletion } = require('./lib/setup-completion');
const { removeCompletion } = require('./lib/remove-completion');
const { installGum, GUM_PATH } = require('./lib/utils/install-gum');

// Try to load embedded scripts. Fallback to empty if not generated yet.
let embedded = { scripts: {}, generalScripts: {} };
try {
  embedded = require('./generated-scripts');
} catch (e) {
  // During initial build, this might be missing
}

const binName = 'minhthetus-cli';

/**
 * Recreates the script structure in a temporary directory to allow sourcing and execution.
 */
function extractScripts() {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), `${binName}-`));
  
  const scriptsDir = path.join(tempDir, 'scripts');
  const genDir = path.join(tempDir, 'generalScripts');
  
  fs.mkdirSync(scriptsDir, { recursive: true });
  fs.mkdirSync(genDir, { recursive: true });

  for (const [relPath, content] of Object.entries(embedded.scripts)) {
    const fullPath = path.join(scriptsDir, relPath);
    fs.mkdirSync(path.dirname(fullPath), { recursive: true });
    fs.writeFileSync(fullPath, content, { mode: 0o755 });
  }

  for (const [relPath, content] of Object.entries(embedded.generalScripts)) {
    const fullPath = path.join(genDir, relPath);
    fs.mkdirSync(path.dirname(fullPath), { recursive: true });
    fs.writeFileSync(fullPath, content, { mode: 0o755 });
  }

  return tempDir;
}

/**
 * Ensures gum is installed and available.
 */
async function ensureGum() {
  if (!fs.existsSync(GUM_PATH)) {
    process.stderr.write("✦ Improving your experience (installing assets)...\n");
    try {
      await installGum();
    } catch (e) {
      console.error("Warning: Failed to install gum. Some UI features may be degraded.");
    }
  }
}

// --- Completion Logic ---
const completion = omelette(binName);

completion.on('complete', (fragment, { reply, line }) => {
  const parts = line.split(/\s+/).filter(Boolean);
  const isNewWord = line.endsWith(' ');
  const args = parts.slice(1, isNewWord ? parts.length : parts.length - 1);
  
  const scriptKeys = Object.keys(embedded.scripts);
  let matches = [];

  if (args.length === 0) {
    matches.push('help', 'setup-completion', 'remove-completion');
  }

  // Find matches based on current args prefix
  const prefix = args.join('/');
  const subItems = new Set();

  for (const key of scriptKeys) {
    if (prefix === '' || key.startsWith(prefix + (prefix ? '/' : ''))) {
      const remaining = prefix === '' ? key : key.slice(prefix.length + 1);
      const firstPart = remaining.split('/')[0];
      if (firstPart) {
        if (firstPart.endsWith('.sh')) {
           subItems.add(firstPart.slice(0, -3));
        } else {
           subItems.add(firstPart);
        }
      }
    }
  }

  reply([...matches, ...Array.from(subItems)]);
});

completion.next(async () => {
  await ensureGum();
  const fullArgs = process.argv.slice(2);

  if (fullArgs.length === 0 || ['help', '--help', '-h'].includes(fullArgs[0])) {
    await showHelp(binName, { embeddedScripts: embedded.scripts });
    process.stderr.write('\n'); // Spacing end 1 line
    process.exit(0);
  }

  if (fullArgs[0] === 'setup-completion') {
    process.stderr.write('\n'); // Spacing top 1 line
    setupCompletion(binName);
    process.stderr.write('\n'); // Spacing end 1 line
    process.exit(0);
  }

  if (fullArgs[0] === 'remove-completion') {
    process.stderr.write('\n'); // Spacing top 1 line
    removeCompletion(binName);
    process.stderr.write('\n'); // Spacing end 1 line
    process.exit(0);
  }

  // Find the script in embedded scripts
  let matchedRelPath = null;
  let consumedCount = 0;

  // Try to find the longest matching path
  for (let i = fullArgs.length; i > 0; i--) {
    const potentialPath = fullArgs.slice(0, i).join('/') + '.sh';
    if (embedded.scripts[potentialPath]) {
      matchedRelPath = potentialPath;
      consumedCount = i;
      break;
    }
  }

  if (!matchedRelPath) {
    process.stderr.write('\n'); // Spacing top 1 line
    console.error(`\x1b[31m\x1b[1mError:\x1b[0m command "${fullArgs.join(' ')}" not found.`);
    process.stderr.write('\n'); // Spacing between error and help
    await showHelp(binName, { skipSplash: true, embeddedScripts: embedded.scripts });
    process.stderr.write('\n'); // Spacing end 1 line
    process.exit(1);
  }

  const scriptArgs = fullArgs.slice(consumedCount);
  
  // Extract scripts to temp dir for execution
  const tempDir = extractScripts();
  const scriptPath = path.join(tempDir, 'scripts', matchedRelPath);

  const binDir = path.dirname(GUM_PATH);
  const newPath = `${binDir}${path.delimiter}${process.env.PATH}`;

  process.stderr.write('\n'); // Spacing top 1 line
  const proc = spawn('sh', [scriptPath, ...scriptArgs], { 
    stdio: 'inherit',
    env: { ...process.env, PATH: newPath, MINHTHETUS_TMP: tempDir } 
  });

  proc.on('exit', code => {
    // Cleanup temp dir
    try {
      fs.rmSync(tempDir, { recursive: true, force: true });
    } catch (e) {}
    process.stderr.write('\n'); // Spacing end 1 line
    process.exit(code || 0);
  });

  proc.on('error', err => {
    console.error(`\x1b[31m\x1b[1mExecution error:\x1b[0m ${err.message}`);
    try {
      fs.rmSync(tempDir, { recursive: true, force: true });
    } catch (e) {}
    process.exit(1);
  });
});

completion.init();
