const fs = require('fs');

/**
 * Emits a shell command to be executed by the parent shell wrapper.
 * This works by writing the command to the temporary file path defined in MINHTHETUS_SHELL_PIPE.
 * 
 * @param {string} cmd - The shell command to execute (e.g., 'export FOO=bar', 'nvm use 20')
 */
function emitShellCommand(cmd) {
  const pipePath = process.env.MINHTHETUS_SHELL_PIPE;
  if (pipePath) {
    try {
      fs.appendFileSync(pipePath, cmd + '\n');
    } catch (err) {
      // Silently fail if we can't write to the pipe, 
      // as it might be a permission issue or the pipe was removed
      console.error(`Failed to emit shell command: ${err.message}`);
    }
  }
}

module.exports = {
  emitShellCommand
};
