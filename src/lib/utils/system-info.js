const path = require('path');

/**
 * Retrieves environment and system information relevant to the CLI's installation and configuration.
 * Centralizes the logic for shell detection and path resolution.
 * 
 * @returns {Object} An object containing shellEnv, isZsh, homeDir, and the shell configFile path.
 */
function getSystemInfo() {
  const shellEnv = process.env.SHELL || '';
  const isZsh = shellEnv.includes('zsh');
  const homeDir = process.env.HOME || process.env.USERPROFILE;
  
  // Determine the appropriate shell configuration file based on the environment
  const configFile = isZsh 
    ? path.join(homeDir, '.zshrc') 
    : path.join(homeDir, '.bash_profile');
  
  return {
    shellEnv,
    isZsh,
    homeDir,
    configFile
  };
}

module.exports = {
  getSystemInfo
};
