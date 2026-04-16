const fs = require('fs');
const path = require('path');

function minifyShContent(content) {
  const lines = content.split('\n');
  return lines
    .map(line => {
      let processed = line;
      if (!processed.trim().startsWith('#!')) {
        processed = processed.replace(/\s#.*$/, '');
      }
      processed = processed.trim();
      if (processed.startsWith('#') && !processed.startsWith('#!')) {
        return null;
      }
      return processed;
    })
    .filter(line => line !== null && line !== '')
    .join('\n');
}

function collectScripts(dir, baseDir, scripts = {}) {
  const items = fs.readdirSync(dir);
  for (const item of items) {
    const fullPath = path.join(dir, item);
    const stat = fs.statSync(fullPath);
    const relPath = path.relative(baseDir, fullPath);

    if (stat.isDirectory()) {
      collectScripts(fullPath, baseDir, scripts);
    } else if (item.endsWith('.sh')) {
      const content = fs.readFileSync(fullPath, 'utf8');
      scripts[relPath] = minifyShContent(content);
    }
  }
  return scripts;
}

const srcDir = path.resolve(__dirname, '../src');
const scriptsDir = path.join(srcDir, 'scripts');
const generalScriptsDir = path.join(srcDir, 'generalScripts');

const allScripts = {};
if (fs.existsSync(scriptsDir)) collectScripts(scriptsDir, scriptsDir, allScripts);

// For generalScripts, we might want to keep the 'generalScripts/' prefix in keys if they are accessed that way
const generalScripts = {};
if (fs.existsSync(generalScriptsDir)) collectScripts(generalScriptsDir, generalScriptsDir, generalScripts);

const output = `// Auto-generated file. Do not edit.
module.exports = {
  scripts: ${JSON.stringify(allScripts, null, 2)},
  generalScripts: ${JSON.stringify(generalScripts, null, 2)}
};
`;

const targetFile = path.join(srcDir, 'generated-scripts.js');
fs.writeFileSync(targetFile, output, 'utf8');
console.log(`Generated ${targetFile} with embedded scripts.`);
