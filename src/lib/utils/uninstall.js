const fs = require('fs');
const path = require('path');
const { getSystemInfo } = require('./system-info');

/**
 * Cleans up shell-related configurations and files created by the CLI.
 * 
 * @param {string} binName - The name of the CLI binary.
 */
function cleanShellConfig(binName) {
  const { configFile, homeDir } = getSystemInfo();
  
  process.stderr.write(`✦ Removing completion and shell integration for ${binName}...\n`);

  // 1. Remove the static completion and wrapper scripts if they exist
  const staticDir = path.join(homeDir, `.${binName}`);
  const staticCompPath = path.join(staticDir, 'completion.sh');
  const staticWrapperPath = path.join(staticDir, 'shell-wrapper.sh');
  
  if (fs.existsSync(staticCompPath)) {
    process.stderr.write(`  Deleting static completion script...\n`);
    fs.unlinkSync(staticCompPath);
  }
  if (fs.existsSync(staticWrapperPath)) {
    process.stderr.write(`  Deleting shell wrapper script...\n`);
    fs.unlinkSync(staticWrapperPath);
  }
  
  // Optionally remove the directory if it's empty
  if (fs.existsSync(staticDir) && fs.readdirSync(staticDir).length === 0) {
    fs.rmdirSync(staticDir);
  }

  // 2. Remove the blocks from the shell config file (.zshrc or .bash_profile)
  if (fs.existsSync(configFile)) {
    let content = fs.readFileSync(configFile, 'utf8');
    
    const compRegex = new RegExp(`# begin ${binName} completion[\\s\\S]*?# end ${binName} completion\\n?`, 'g');
    const wrapperRegex = new RegExp(`# begin ${binName} shell wrapper[\\s\\S]*?# end ${binName} shell wrapper\\n?`, 'g');
    
    let modified = false;
    if (compRegex.test(content)) {
      content = content.replace(compRegex, '');
      modified = true;
    }
    if (wrapperRegex.test(content)) {
      content = content.replace(wrapperRegex, '');
      modified = true;
    }

    if (modified) {
      fs.writeFileSync(configFile, content.trim() + '\n');
      process.stderr.write(`✅ Shell integration blocks removed from ${configFile}.\n`);
    } else {
      process.stderr.write(`ℹ️  No shell integration blocks found in ${configFile}.\n`);
    }
  } else {
    process.stderr.write(`ℹ️  Shell configuration file not found: ${configFile}. Skipping cleanup.\n`);
  }
}

/**
 * Performs a full uninstallation of the CLI tool.
 * This includes removing shell completion and shell wrappers.
 * 
 * @param {string} binName - The name of the CLI binary.
 */
async function uninstall(binName) {
  process.stderr.write(`✦ Starting uninstallation cleanup for ${binName}...\n`);

  try {
    cleanShellConfig(binName);
  } catch (err) {
    process.stderr.write(`⚠️ Warning: Error during shell cleanup: ${err.message}\n`);
  }

  process.stderr.write(`✅ Shell configuration and completions cleaned up.\n`);
  process.stderr.write(`\n✅ Uninstallation complete. Please restart your terminal or source your shell configuration.\n`);
}

// Support running as a standalone script (useful for preuninstall hook)
if (require.main === module) {
  const binName = 'minhthetus-cli'; // Default binary name
  uninstall(binName).catch((err) => {
    console.error('Uninstall script failed:', err);
    process.exit(1);
  });
}

module.exports = { uninstall };
