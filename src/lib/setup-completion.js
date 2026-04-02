const fs = require('fs');
const path = require('path');

function setupCompletion(binName, completion) {
  console.log(`Setting up completion for ${binName}...`);
  try {
    const shellEnv = process.env.SHELL || '';
    const isZsh = shellEnv.includes('zsh');
    const configFile = isZsh ? path.join(process.env.HOME, '.zshrc') : path.join(process.env.HOME, '.bash_profile');

    if (fs.existsSync(configFile)) {
      let content = fs.readFileSync(configFile, 'utf8');
      if (isZsh && !content.includes('compinit')) {
          content = 'autoload -Uz compinit && compinit\n' + content;
      }
      content = content.replace(/# begin minhthetus-cli completion[\s\S]*?# end minhthetus-cli completion/g, '');
      const block = `\n# begin ${binName} completion\n. <(${binName} --completion)\n# end ${binName} completion\n`;
      fs.writeFileSync(configFile, content.trim() + block);
      console.log(`\n✅ Success! Completion added to ${configFile}. Please run: source ${configFile} or Open new terminal`);
    } else {
      completion.setupShellInitFile();
    }
  } catch (err) {
    console.error('Setup failed:', err.message);
  }
}

module.exports = { setupCompletion };
