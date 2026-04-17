const fs = require('fs');
const path = require('path');

// Colors for terminal output
const RED = '\x1b[31m';
const YELLOW = '\x1b[33m';
const RESET = '\x1b[0m';
const BOLD = '\x1b[1m';

const srcDir = path.resolve(__dirname, '../src');
const scriptsDir = path.join(srcDir, 'scripts');
const generalScriptsDir = path.join(srcDir, 'generalScripts');

const isStrict = process.argv.includes('--strict');
let hasErrors = false;
let foundIssues = false;

function checkScripts(dir) {
  if (!fs.existsSync(dir)) return;
  
  const items = fs.readdirSync(dir);
  for (const item of items) {
    const fullPath = path.join(dir, item);
    const stat = fs.statSync(fullPath);

    if (stat.isDirectory()) {
      checkScripts(fullPath);
    } else if (item.endsWith('.sh')) {
      const content = fs.readFileSync(fullPath, 'utf8');
      if (!content.includes('# Description:')) {
        foundIssues = true;
        const relPath = path.relative(path.resolve(__dirname, '..'), fullPath);
        if (isStrict) {
            console.error(`${RED}${BOLD}Error:${RESET} Script "${relPath}" is missing the "${BOLD}# Description:${RESET}" header.`);
            hasErrors = true;
        } else {
            console.warn(`${YELLOW}${BOLD}Warning:${RESET} Script "${relPath}" is missing the "${BOLD}# Description:${RESET}" header.`);
        }
      }
    }
  }
}

console.log(`${BOLD}Running pre-check for shell scripts...${RESET}`);

checkScripts(scriptsDir);
checkScripts(generalScriptsDir);

if (!foundIssues) {
  console.log(`${BOLD}All shell scripts passed validation.${RESET}`);
} else if (hasErrors) {
  console.error(`${RED}${BOLD}Pre-check failed. Please add descriptions to the scripts listed above.${RESET}`);
  process.exit(1);
} else {
  console.log(`${YELLOW}${BOLD}Pre-check finished with warnings.${RESET}`);
}
