const fs = require('fs');
const path = require('path');

const { showSplash, trackLines, resetLines } = require('./splash');

async function showHelp(binName) {
  resetLines();
  const splashPromise = showSplash(false);
  
  function log(msg = '') {
    console.log(msg);
    // Count lines in message + 1 for console.log's implicit newline
    const count = (msg.match(/\n/g) || []).length + 1;
    trackLines(count);
  }
  
  // Try to find the scripts directory. 
  // When running from src/, it's ../scripts
  // When running from dist/, it's ../src/scripts
  let scriptsDir = path.join(__dirname, '..', 'scripts');
  if (!fs.existsSync(path.join(scriptsDir, 'hello.sh')) && fs.existsSync(path.join(__dirname, '..', 'src', 'scripts'))) {
    scriptsDir = path.join(__dirname, '..', 'src', 'scripts');
  }
  
  const colors = {
    reset: "\x1b[0m",
    bright: "\x1b[1m",
    blue: "\x1b[34m",
    green: "\x1b[32m",
    cyan: "\x1b[36m",
    gray: "\x1b[90m"
  };

  log(`\n${colors.bright}Usage:${colors.reset} ${binName} ${colors.cyan}<command>${colors.reset} [args]`);
  log(`\n${colors.bright}Available commands:${colors.reset}`);

  function getScriptsTree(dir) {
    if (!fs.existsSync(dir)) return {};
    const tree = {};
    const items = fs.readdirSync(dir);

    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stat = fs.statSync(fullPath);
      if (stat.isDirectory()) {
        const subTree = getScriptsTree(fullPath);
        if (Object.keys(subTree).length > 0) {
          tree[item] = { type: 'dir', children: subTree };
        }
      } else if (item.endsWith('.sh')) {
        const name = item.slice(0, -3);
        const scriptContent = fs.readFileSync(fullPath, 'utf8');
        const match = scriptContent.match(/#\s*Description:\s*(.+)/i);
        tree[name] = { 
          type: 'file', 
          description: match ? match[1].trim() : 'No description provided' 
        };
      }
    }
    return tree;
  }

  function printTree(tree, indent = '', isLast = true) {
    const keys = Object.keys(tree).sort();
    
    keys.forEach((key, index) => {
      const isLastItem = index === keys.length - 1;
      const item = tree[key];
      const connector = isLastItem ? '└── ' : '├── ';
      
      if (item.type === 'dir') {
        log(`${indent}${connector}${colors.blue}${colors.bright}${key}/${colors.reset}`);
        printTree(item.children, indent + (isLastItem ? '    ' : '│   '), true);
      } else {
        const description = item.description ? ` ${colors.gray}# ${item.description}${colors.reset}` : '';
        log(`${indent}${connector}${colors.green}${key}${colors.reset}${description}`);
      }
    });
  }

  const scriptsTree = getScriptsTree(scriptsDir);
  printTree(scriptsTree);
  log();

  // Built-in commands
  log(`${colors.bright}Built-in commands:${colors.reset}`);
  log(`  ${colors.cyan}help${colors.reset}               Show this help message`);
  log(`  ${colors.cyan}setup-completion${colors.reset}   Install tab completion for your shell`);
  log(`  ${colors.cyan}remove-completion${colors.reset}  Uninstall tab completion for your shell\n`);

  await splashPromise;
}

module.exports = { showHelp };
