const fs = require('fs');
const path = require('path');

function removeCompletion(binName) {
  console.log(`Removing completion for ${binName}...`);
  try {
    const shellEnv = process.env.SHELL || '';
    const isZsh = shellEnv.includes('zsh');
    const homeDir = process.env.HOME || process.env.USERPROFILE;
    const configFile = isZsh ? path.join(homeDir, '.zshrc') : path.join(homeDir, '.bash_profile');

    // 1. Remove the static completion script if it exists
    const staticCompPath = path.join(homeDir, `.${binName}-completion.sh`);
    if (fs.existsSync(staticCompPath)) {
      console.log(`Deleting static completion script at ${staticCompPath}...`);
      fs.unlinkSync(staticCompPath);
    }

    // 2. Remove the block from the shell config file
    if (fs.existsSync(configFile)) {
      let content = fs.readFileSync(configFile, 'utf8');
      
      const regex = new RegExp(`# begin ${binName} completion[\\s\\S]*?# end ${binName} completion\\n?`, 'g');
      
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
