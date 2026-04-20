const { installGum } = require('./install-gum');
const { setupCompletion } = require('../setup-completion');
const { setupShellWrapper } = require('./setup-shell-wrapper');
const { emitShellCommand } = require('./emit-shell-command');
const { getSystemInfo } = require('./system-info');
const path = require('path');

/**
 * Orchestrates the installation of all necessary dependencies and configurations.
 * Runs gum binary installation and shell completion setup in parallel for efficiency.
 */
async function main() {
  console.log('✦ Initializing minhthetus-cli...');
  
  const binName = 'minhthetus-cli';
  // Point to the entry point so setupCompletion can extract the completion script
  // from the main index.js logic.
  const scriptPath = path.join(__dirname, '../../index.js');

  try {
    // Run installation tasks in parallel to speed up the postinstall process.
    await Promise.all([
      installGum().catch(err => {
        console.error('✖ Failed to install gum:', err.message);
      }),
      (async () => {
        try {
          setupCompletion(binName, scriptPath);
        } catch (err) {
          console.error('✖ Failed to setup completion:', err.message);
        }
      })(),
      (async () => {
        try {
          setupShellWrapper(binName);
        } catch (err) {
          console.error('✖ Failed to setup shell wrapper:', err.message);
        }
      })()
    ]);

    
    
    console.log('\n✦ All dependencies and configurations are resolved.');

    // If running via the shell wrapper, emit a command to source the config file
    // so the changes take effect immediately in the current session.
    const { configFile } = getSystemInfo();

    if (process.env.MINHTHETUS_SHELL_PIPE) {
      console.log(`✦ Emitting 'source ${configFile}' to current shell...`);
      emitShellCommand(`source "${configFile}"`);
    } else {
      console.log(`\n👉 Please run: "source ${configFile}" to apply changes to this terminal.`);
    }
  } catch (err) {
    console.error('An unexpected error occurred during installation:', err);
  }
}

if (require.main === module) {
  main();
}
