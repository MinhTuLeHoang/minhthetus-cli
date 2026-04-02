#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');
const omelette = require('omelette');

const { showHelp } = require('./lib/help');
const { setupCompletion } = require('./lib/setup-completion');
const { removeCompletion } = require('./lib/remove-completion');

const binName = 'minhthetus-cli';
const scriptsDirRoot = path.join(__dirname, 'scripts');

// --- Completion Logic ---
const completion = omelette(binName);

completion.on('complete', (fragment, { reply, line }) => {
  // Use regex to split and filter out empty strings from trailing spaces
  const parts = line.split(/\s+/).filter(Boolean);
  // args are words after binName, excluding the word currently being typed
  // If line ends with space, then we are completing a NEW word.
  // If line does NOT end with space, then we are completing the LAST word.
  const isNewWord = line.endsWith(' ');
  const args = parts.slice(1, isNewWord ? parts.length : parts.length - 1);
  
  let currentDir = scriptsDirRoot;
  let matches = [];

  // If this is the first argument, include built-in commands
  if (args.length === 0) {
    matches.push('help', 'setup-completion', 'remove-completion');
  }

  // Navigate through subfolders based on existing arguments
  for (const arg of args) {
    if (['help', 'setup-completion', 'remove-completion', 'hello'].includes(arg)) {
       return reply([]); 
    }
    const nextDir = path.join(currentDir, arg);
    if (fs.existsSync(nextDir) && fs.statSync(nextDir).isDirectory()) {
      currentDir = nextDir;
    } else {
      // Hit a file or invalid path
      return reply([]);
    }
  }

  // Find available items in currentDir
  if (fs.existsSync(currentDir) && fs.statSync(currentDir).isDirectory()) {
    const items = fs.readdirSync(currentDir);
    for (const item of items) {
      const fullPath = path.join(currentDir, item);
      const stat = fs.statSync(fullPath);
      if (stat.isDirectory()) {
        matches.push(item);
      } else if (item.endsWith('.sh')) {
        matches.push(item.slice(0, -3));
      }
    }
  }

  reply(matches);
});

completion.next(() => {
  // --- Execution Logic ---
  const fullArgs = process.argv.slice(2);

  if (fullArgs.length === 0 || ['help', '--help', '-h'].includes(fullArgs[0])) {
    showHelp(binName);
    process.exit(0);
  }

  if (fullArgs[0] === 'setup-completion') {
    setupCompletion(binName, completion);
    process.exit(0);
  }

  if (fullArgs[0] === 'remove-completion') {
    removeCompletion(binName);
    process.exit(0);
  }

  // Find the script path by consuming arguments
  let scriptPath = null;
  let currentSearchDir = scriptsDirRoot;
  let consumedCount = 0;

  for (let i = 0; i < fullArgs.length; i++) {
    const arg = fullArgs[i];
    const potentialFile = path.join(currentSearchDir, `${arg}.sh`);
    const potentialDir = path.join(currentSearchDir, arg);

    if (fs.existsSync(potentialFile) && fs.statSync(potentialFile).isFile()) {
      scriptPath = potentialFile;
      consumedCount = i + 1;
      break;
    } else if (fs.existsSync(potentialDir) && fs.statSync(potentialDir).isDirectory()) {
      currentSearchDir = potentialDir;
    } else {
      break;
    }
  }

  if (!scriptPath) {
    console.error(`\nError: command "${fullArgs.join(' ')}" not found.`);
    showHelp(binName);
    process.exit(1);
  }

  const scriptArgs = fullArgs.slice(consumedCount);

  spawn('sh', [scriptPath, ...scriptArgs], { stdio: 'inherit' })
    .on('exit', code => process.exit(code || 0))
    .on('error', err => {
      console.error('Execution error:', err.message);
      process.exit(1);
    });
});

completion.init();

