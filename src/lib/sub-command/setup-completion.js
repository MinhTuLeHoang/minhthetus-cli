const { setupCompletion } = require('../setup-completion');
const { setupShellWrapper } = require('../utils/setup-shell-wrapper');

/**
 * Handles the 'setup-completion' sub-command.
 * @param {string} binName 
 */
module.exports = (binName) => {
  process.stderr.write('\n'); // Spacing top 1 line
  setupCompletion(binName);
  setupShellWrapper(binName);
  process.stderr.write('\n'); // Spacing end 1 line
  process.exit(0);
};
