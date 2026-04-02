const fs = require('fs');
const path = require('path');

function showHelp(binName) {
  console.log(`\nUsage: ${binName} <command> [args]`);
  console.log('\nAvailable commands:');

  const scriptsDir = path.join(__dirname, '..', 'scripts');
  const scripts = [];

  function findScripts(dir, prefix = '') {
    if (!fs.existsSync(dir)) return;
    const items = fs.readdirSync(dir);
    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stat = fs.statSync(fullPath);
      if (stat.isDirectory()) {
        findScripts(fullPath, prefix + item + ' ');
      } else if (item.endsWith('.sh')) {
        const name = prefix + item.slice(0, -3);
        const scriptContent = fs.readFileSync(fullPath, 'utf8');
        const match = scriptContent.match(/#\s*Description:\s*(.+)/i);
        const description = match ? match[1].trim() : '';
        scripts.push({ name, description });
      }
    }
  }

  findScripts(scriptsDir);

  const maxLen = Math.max(0, ...scripts.map(s => s.name.length), 16);
  scripts.sort((a, b) => a.name.localeCompare(b.name)).forEach(s => {
    console.log(`  - ${s.name.padEnd(maxLen)}   ${s.description}`);
  });

  console.log(`  - ${'help'.padEnd(maxLen)}   Show this help message`);
  console.log(`  - ${'setup-completion'.padEnd(maxLen)}   Install tab completion for your shell`);
  console.log(`  - ${'remove-completion'.padEnd(maxLen)}   Uninstall tab completion for your shell\n`);
}

module.exports = { showHelp };

