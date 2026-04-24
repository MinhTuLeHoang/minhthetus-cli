const { uninstall } = require('../utils/uninstall');
const { execSync } = require('child_process');

/**
 * Handles the 'uninstall' sub-command.
 * @param {string} binName 
 */
module.exports = async (binName) => {
  process.stderr.write('\n'); // Spacing top 1 line
  
  // First, perform the internal cleanup (completions, shell wrappers, etc.)
  await uninstall(binName);
  
  process.stderr.write(`\n✦ Attempting to remove the package globally...\n`);
  
  // Try pnpm first, then npm. Standard error is ignored to keep it clean if one fails.
  try {
    try {
      process.stderr.write(`Running: pnpm uninstall -g ${binName}\n`);
      execSync(`pnpm uninstall -g ${binName}`, { stdio: 'inherit' });
    } catch (e) {
      process.stderr.write(`Running: npm uninstall -g ${binName}\n`);
      execSync(`npm uninstall -g ${binName}`, { stdio: 'inherit' });
    }
    process.stderr.write(`\n✅ Package removed successfully.\n`);
  } catch (err) {
    process.stderr.write(`\n⚠️ Could not automatically remove the package.\n`);
    process.stderr.write(`Please run one of the following manually to finish:\n`);
    process.stderr.write(`   pnpm uninstall -g ${binName}\n`);
    process.stderr.write(`   npm uninstall -g ${binName}\n`);
  }
  
  process.stderr.write('\n'); // Spacing end 1 line
  process.exit(0);
};
