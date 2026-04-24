const { addRepo } = require('../utils/repo-config');

/**
 * Handles the 'repo-track' sub-command.
 * @param {string} binName 
 * @param {string[]} args 
 */
module.exports = (binName, args) => {
  const isSilent = process.argv.includes('--silent');
  if (args[1]) {
    addRepo(args[1], { silent: isSilent });
  }
  process.exit(0);
};
