const fs = require('fs');
const path = require('path');

function removeCompletion(binName) {
  console.log(`Removing completion for ${binName}...`);
  try {
    const shellEnv = process.env.SHELL || '';
    const isZsh = shellEnv.includes('zsh');
    const homeDir = process.env.HOME || process.env.USERPROFILE;
    const configFile = isZsh ? path.join(homeDir, '.zshrc') : path.join(homeDir, '.bash_profile');

    // 1. Remove the static completion and wrapper scripts if they exist
    const staticDir = path.join(homeDir, `.${binName}`);
    const staticCompPath = path.join(staticDir, 'completion.sh');
    const staticWrapperPath = path.join(staticDir, 'shell-wrapper.sh');
    
    if (fs.existsSync(staticCompPath)) {
      console.log(`Deleting static completion script at ${staticCompPath}...`);
      fs.unlinkSync(staticCompPath);
    }
    if (fs.existsSync(staticWrapperPath)) {
      console.log(`Deleting shell wrapper script at ${staticWrapperPath}...`);
      fs.unlinkSync(staticWrapperPath);
    }
    
    // Optionally remove the directory if it's empty
    if (fs.existsSync(staticDir) && fs.readdirSync(staticDir).length === 0) {
      fs.rmdirSync(staticDir);
    }

    // 2. Remove the blocks from the shell config file
    if (fs.existsSync(configFile)) {
      let content = fs.readFileSync(configFile, 'utf8');
      
      const compRegex = new RegExp(`# begin ${binName} completion[\\s\\S]*?# end ${binName} completion\\n?`, 'g');
      const wrapperRegex = new RegExp(`# begin ${binName} shell wrapper[\\s\\S]*?# end ${binName} shell wrapper\\n?`, 'g');
      
      let modified = false;
      if (compRegex.test(content)) {
        content = content.replace(compRegex, '');
        modified = true;
      }
      if (wrapperRegex.test(content)) {
        content = content.replace(wrapperRegex, '');
        modified = true;
      }

      if (modified) {
        fs.writeFileSync(configFile, content.trim() + '\n');
        console.log(`\n✅ Success! Shell integration removed from ${configFile}. Please run: source ${configFile} or restart your terminal.`);
      } else {
        console.log(`\nℹ️  No shell integration blocks found for ${binName} in ${configFile}.`);
      }
    } else {
      console.error(`\n❌ Error: Shell config file not found: ${configFile}`);
    }
  } catch (err) {
    console.error('\n❌ Removal failed:', err.message);
  }
}

module.exports = { removeCompletion };
