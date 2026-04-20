const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

function setupCompletion(binName) {
  console.log(`Setting up completion for ${binName}...`);
  try {
    const shellEnv = process.env.SHELL || '';
    const isZsh = shellEnv.includes('zsh');
    const homeDir = process.env.HOME || process.env.USERPROFILE;
    const configFile = isZsh ? path.join(homeDir, '.zshrc') : path.join(homeDir, '.bash_profile');
    
    // 1. Generate and save the static completion script
    const staticCompDir = path.join(homeDir, `.${binName}`);
    const staticCompPath = path.join(staticCompDir, 'completion.sh');
    
    if (!fs.existsSync(staticCompDir)) {
      fs.mkdirSync(staticCompDir, { recursive: true });
    }

    console.log(`Generating static completion script at ${staticCompPath}...`);
    
    // Safely invoke the local CLI using the Node executable to get the completion string
    const scriptPath = path.join(__dirname, '..', 'index.js');
    const completionOutput = execSync(`"${process.execPath}" "${scriptPath}" --completion`, { encoding: 'utf8' });
    fs.writeFileSync(staticCompPath, completionOutput);

    // 2. Modify the user's rc file to source the newly created static script
    if (fs.existsSync(configFile)) {
      let content = fs.readFileSync(configFile, 'utf8');
      if (isZsh && !content.includes('compinit')) {
          content = 'autoload -Uz compinit && compinit\n' + content;
      }
      
      // Remove old dynamic/slow completion block if it exists
      content = content.replace(new RegExp(`# begin ${binName} completion[\\s\\S]*?# end ${binName} completion\\n?`, 'g'), '');
      
      // Append the new, fast static sourcing block
      const block = `\n# begin ${binName} completion\n[ -f "${staticCompPath}" ] && source "${staticCompPath}"\n# end ${binName} completion\n`;
      
      fs.writeFileSync(configFile, content.trim() + '\n' + block);
      console.log(`\n✅ Success! Completion added to ${configFile}.`);
      console.log(`Please run: "source ${configFile}" or Open a new terminal`);
    } else {
      console.error(`Shell config file not found at ${configFile}`);
    }
  } catch (err) {
    console.error('Setup failed:', err.message);
  }
}

module.exports = { setupCompletion };
