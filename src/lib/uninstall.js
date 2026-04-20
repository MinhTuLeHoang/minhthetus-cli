const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const { removeCompletion } = require('./remove-completion');
const { getSystemInfo } = require('./utils/system-info');

/**
 * Performs a full uninstallation of the CLI tool.
 * This includes removing shell completion, shell wrappers, and the global binary.
 * 
 * @param {string} binName - The name of the CLI binary.
 */
async function uninstall(binName) {
  const { homeDir } = getSystemInfo();
  
  process.stderr.write(`✦ Starting uninstallation of ${binName}...\n`);

  // 1. Remove shell completion and shell wrapper
  // The removeCompletion function handles deleting static scripts and removing shell rc blocks.
  try {
    removeCompletion(binName);
  } catch (err) {
    process.stderr.write(`⚠️ Warning: Error during completion removal: ${err.message}\n`);
  }

  // 2. Delete the binary in pnpm global bin directory
  // We attempt to locate the pnpm global bin directory to remove the executable.
  try {
    const possibleBinPaths = new Set();
    
    // Attempt to get the global-bin-dir from pnpm config
    try {
      const pnpmBinDir = execSync('pnpm config get global-bin-dir', { encoding: 'utf8' }).trim();
      if (pnpmBinDir && pnpmBinDir !== 'undefined' && pnpmBinDir !== 'null') {
        possibleBinPaths.add(path.join(pnpmBinDir, binName));
      }
    } catch (e) {
      // pnpm config might not be available or fails
    }

    // Attempt to get bin path via 'pnpm bin -g'
    try {
      const pnpmBinDir = execSync('pnpm bin -g', { encoding: 'utf8' }).trim();
      if (pnpmBinDir) {
        possibleBinPaths.add(path.join(pnpmBinDir, binName));
      }
    } catch (e) {
      // pnpm bin -g might fail if not in a pnpm context or command not found
    }

    // Common Mac/Linux paths for pnpm global binaries
    possibleBinPaths.add(path.join(homeDir, '.local/share/pnpm', binName));
    possibleBinPaths.add(path.join(homeDir, 'Library/pnpm', binName));
    
    let removed = false;
    for (const binFile of possibleBinPaths) {
      if (fs.existsSync(binFile)) {
        process.stderr.write(`Removing global binary at ${binFile}...\n`);
        fs.unlinkSync(binFile);
        removed = true;
      }
    }

    if (removed) {
      process.stderr.write(`✅ Global binary successfully removed.\n`);
    } else {
      process.stderr.write(`ℹ️  Note: Global binary not found in standard pnpm locations.\n`);
      process.stderr.write(`   If you installed via another package manager, please uninstall manually (e.g., npm uninstall -g ${binName}).\n`);
    }
  } catch (err) {
    process.stderr.write(`⚠️ Warning: Failed to remove global binary: ${err.message}\n`);
  }

  process.stderr.write(`\n✅ Uninstallation complete. Please restart your terminal or source your shell configuration.\n`);
}

module.exports = { uninstall };
