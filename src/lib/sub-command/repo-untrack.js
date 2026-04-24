const { removeRepo } = require('../utils/repo-config');

/**
 * Handles the 'repo-untrack' sub-command.
 * @param {string} binName 
 * @param {string[]} args 
 */
module.exports = (binName, args) => {
  if (args[1]) {
    removeRepo(args[1]);
  }
  process.exit(0);
};
