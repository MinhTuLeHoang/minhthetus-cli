const fs = require('fs');
const path = require('path');

const { showSplash, trackLines, resetLines } = require('./utils/splash');
const { GUM_PATH } = require('./utils/install-gum');
const { execSync } = require('child_process');

function gumStyle(text, options = {}) {
  if (!fs.existsSync(GUM_PATH)) return null;
  const args = Object.entries(options).map(([k, v]) => `--${k} "${v}"`).join(' ');
  try {
    return execSync(`${GUM_PATH} style ${args} "${text}"`).toString();
  } catch (e) {
    return null;
  }
}

async function showHelp(binName, { skipSplash = false, embeddedScripts = null, remoteRegistry = null } = {}) {
  resetLines();
  let splashPromise = null;
  if (!skipSplash) {
    splashPromise = showSplash(false);
  }
  
  function log(msg = '') {
    process.stderr.write(msg + '\n');
    const count = (msg.match(/\n/g) || []).length + 1;
    if (!skipSplash) trackLines(count);
  }
  
  const colors = {
    reset: "\x1b[0m",
    bright: "\x1b[1m",
    blue: "\x1b[34m",
    green: "\x1b[32m",
    cyan: "\x1b[36m",
    magenta: "\x1b[35m",
    gray: "\x1b[90m"
  };

  const usageHeader = gumStyle("USAGE", { foreground: "212", border: "normal", padding: "0 1", "border-foreground": "212" });
  if (usageHeader) {
    log(usageHeader.trim());
    log(`${colors.bright}${binName}${colors.reset} ${colors.cyan}<command>${colors.reset} [args]`);
  } else {
    log(`${colors.bright}Usage:${colors.reset} ${binName} ${colors.cyan}<command>${colors.reset} [args]`);
  }

  // spacing 1 line
  log("");
  
  const commandsHeader = gumStyle("AVAILABLE COMMANDS", { foreground: "99", margin: "1 0 0 0" });
  if (commandsHeader) {
    log(commandsHeader.trim());
  } else {
    log(`\n${colors.bright}Available commands:${colors.reset}`);
  }

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

  // Merge Remote Registry
  if (remoteRegistry) {
    for (const [relPath, info] of Object.entries(remoteRegistry)) {
      const parts = relPath.split('/');
      let current = scriptsTree;
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i];
        if (i === parts.length - 1) {
          if (!current[part]) {
            current[part] = {
              type: 'file',
              description: (info.description || 'Remote command'),
              isRemote: true
            };
          }
        } else {
          if (!current[part]) current[part] = { type: 'dir', children: {} };
          current = current[part].children;
        }
      }
    }
  }

  // Merge Remote Scripts on Disk (if any)
  const remoteDir = path.join(__dirname, '..', 'remote-scripts');
  if (fs.existsSync(remoteDir)) {
    const addRemoteOnDisk = (dir, currentTree) => {
      fs.readdirSync(dir).forEach(item => {
        const full = path.join(dir, item);
        if (fs.statSync(full).isDirectory()) {
          if (!currentTree[item]) currentTree[item] = { type: 'dir', children: {} };
          addRemoteOnDisk(full, currentTree[item].children);
        } else if (item.endsWith('.sh')) {
          const name = item.slice(0, -3);
          const content = fs.readFileSync(full, 'utf8');
          const match = content.match(/#\s*Description:\s*(.+)/i);
          currentTree[name] = {
            type: 'file',
            description: (match ? match[1].trim() : 'Remote script'),
            isRemote: true
          };
        }
      });
    };
    addRemoteOnDisk(remoteDir, scriptsTree);
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
        const color = item.isRemote ? colors.magenta : colors.green;
        log(`${indent}${connector}${color}${key}${colors.reset}${description}`);
      }
    });
  }

  printTree(scriptsTree);
  log();
  log(`${colors.gray}Legend: ${colors.blue}Modules/${colors.reset}${colors.gray}, ${colors.green}Core Commands${colors.reset}${colors.gray}, ${colors.magenta}Remote Scripts (via npx)${colors.reset}`);
  log();

  // Built-in commands
  log(`${colors.bright}Built-in commands:${colors.reset}`);
  log(`  ${colors.cyan}help${colors.reset}               Show this help message`);
  log(`  ${colors.cyan}setup-completion${colors.reset}   Install tab completion for your shell`);
  log(`  ${colors.cyan}uninstall${colors.reset}          Completely remove the CLI and all integrations`);

  if (splashPromise) await splashPromise;
}

module.exports = { showHelp };
