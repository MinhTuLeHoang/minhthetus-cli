const fs = require('fs');
const path = require('path');

const { showSplash, trackLines, resetLines } = require('./splash');

async function showHelp(binName, { skipSplash = false, embeddedScripts = null } = {}) {
  resetLines();
  let splashPromise = null;
  if (!skipSplash) {
    splashPromise = showSplash(false);
  }
  
  function log(msg = '') {
    console.log(msg);
    const count = (msg.match(/\n/g) || []).length + 1;
    if (!skipSplash) trackLines(count);
  }
  
  const colors = {
    reset: "\x1b[0m",
    bright: "\x1b[1m",
    blue: "\x1b[34m",
    green: "\x1b[32m",
    cyan: "\x1b[36m",
    gray: "\x1b[90m"
  };

  log(`${colors.bright}Usage:${colors.reset} ${binName} ${colors.cyan}<command>${colors.reset} [args]`);
  log(`\n${colors.bright}Available commands:${colors.reset}`);

  let scriptsTree = {};

  if (embeddedScripts) {
    // Build tree from embedded scripts
    for (const [relPath, content] of Object.entries(embeddedScripts)) {
      const parts = relPath.split('/');
      let current = scriptsTree;
      
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i];
        const isLast = i === parts.length - 1;
        
        if (isLast) {
          const name = part.endsWith('.sh') ? part.slice(0, -3) : part;
          const match = content.match(/#\s*Description:\s*(.+)/i);
          current[name] = {
            type: 'file',
            description: match ? match[1].trim() : 'No description provided'
          };
        } else {
          if (!current[part]) {
            current[part] = { type: 'dir', children: {} };
          }
          current = current[part].children;
        }
      }
    }
  } else {
    // Fallback to filesystem
    let scriptsDir = path.join(__dirname, '..', 'scripts');
    if (!fs.existsSync(path.join(scriptsDir, 'hello.sh')) && fs.existsSync(path.join(__dirname, '..', 'src', 'scripts'))) {
      scriptsDir = path.join(__dirname, '..', 'src', 'scripts');
    }

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
    scriptsTree = getScriptsTree(scriptsDir);
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

  printTree(scriptsTree);
  log();

  // Built-in commands
  log(`${colors.bright}Built-in commands:${colors.reset}`);
  log(`  ${colors.cyan}help${colors.reset}               Show this help message`);
  log(`  ${colors.cyan}setup-completion${colors.reset}   Install tab completion for your shell`);
  log(`  ${colors.cyan}remove-completion${colors.reset}  Uninstall tab completion for your shell`);

  if (splashPromise) await splashPromise;
}

module.exports = { showHelp };
