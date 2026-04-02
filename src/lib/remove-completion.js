const fs = require('fs');
const path = require('path');

function removeCompletion(binName) {
  console.log(`Removing completion for ${binName}...`);
  try {
    const shellEnv = process.env.SHELL || '';
    const isZsh = shellEnv.includes('zsh');
    const configFile = isZsh ? path.join(process.env.HOME, '.zshrc') : path.join(process.env.HOME, '.bash_profile');

    if (fs.existsSync(configFile)) {
      let content = fs.readFileSync(configFile, 'utf8');
      
      const regex = new RegExp(`# begin ${binName} completion[\\s\\S]*?# end ${binName} completion`, 'g');
      
      if (regex.test(content)) {
        content = content.replace(regex, '').trim();
        fs.writeFileSync(configFile, content + '\n');
        console.log(`\n✅ Success! Completion block removed from ${configFile}. Please run: source ${configFile} or restart your terminal.`);
      } else {
        console.log(`\nℹ️  No completion block found for ${binName} in ${configFile}.`);
      }
    } else {
      console.error(`\n❌ Error: Shell config file not found: ${configFile}`);
    }
  } catch (err) {
    console.error('\n❌ Removal failed:', err.message);
  }
}

module.exports = { removeCompletion };
