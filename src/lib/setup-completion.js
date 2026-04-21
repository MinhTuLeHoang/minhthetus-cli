const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const { emitShellCommand } = require('./utils/emit-shell-command');
const { getSystemInfo } = require('./utils/system-info');

function setupCompletion(binName, scriptPath) {
  console.log(`Setting up completion for ${binName}...`);
  try {
    const { isZsh, homeDir, configFile } = getSystemInfo();
    
    // 1. Generate and save the static completion script
    const staticCompDir = path.join(homeDir, `.${binName}`);
    const staticCompPath = path.join(staticCompDir, 'completion.sh');
    
    if (!fs.existsSync(staticCompDir)) {
      fs.mkdirSync(staticCompDir, { recursive: true });
    }

    console.log(`Generating static completion script at ${staticCompPath}...`);
    
    // Safely invoke the local CLI using the Node executable to get the completion string
    const targetScript = scriptPath || (require.main ? require.main.filename : process.argv[1]);
    const completionOutput = execSync(`"${process.execPath}" "${targetScript}" --completion`, { encoding: 'utf8' });
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
      const block = `
# begin ${binName} completion
[ -f "${staticCompPath}" ] && source "${staticCompPath}"
# end ${binName} completion
`;
      
      fs.writeFileSync(configFile, content.trim() + '\n' + block);
      console.log(`\n✅ Success! Completion added to ${configFile}.`);
      
      if (process.env.MINHTHETUS_SHELL_PIPE) {
        emitShellCommand(`source "${configFile}"`);
      } else {
        console.log(`Please run: "source ${configFile}" or Open a new terminal`);
      }
    } else {
      console.error(`Shell config file not found at ${configFile}`);
    }
  } catch (err) {
    console.error('Setup failed:', err.message);
  }
}

module.exports = { setupCompletion };
