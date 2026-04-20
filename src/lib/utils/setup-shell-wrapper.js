const fs = require('fs');
const path = require('path');
const { emitShellCommand } = require('./emit-shell-command');
const { getSystemInfo } = require('./system-info');

/**
 * Sets up a shell function wrapper that allows the CLI to modify the parent shell environment.
 * It creates a static script and adds a sourcing line to the user's shell configuration file.
 * 
 * @param {string} binName - The name of the CLI binary (e.g., 'minhthetus-cli').
 */
function setupShellWrapper(binName) {
  console.log(`✦ Setting up shell wrapper for ${binName}...`);
  try {
    const { isZsh, homeDir, configFile } = getSystemInfo();
    
    // 1. Generate and save the static shell wrapper script
    const staticDir = path.join(homeDir, `.${binName}`);
    const wrapperPath = path.join(staticDir, 'shell-wrapper.sh');
    
    if (!fs.existsSync(staticDir)) {
      fs.mkdirSync(staticDir, { recursive: true });
    }

    console.log(`Generating shell wrapper script at ${wrapperPath}...`);
    
    // The wrapper function captures environment modifications emitted by the CLI
    const wrapperContent = `
# Shell wrapper for ${binName} to enable environment modifications (e.g. nvm use)
${binName}() {
  # Bypass the wrapper for completion requests to avoid overhead and potential issues
  if [[ "$*" == *"--compgen"* || "$*" == *"--compzsh"* || "$*" == *"--compbash"* ]]; then
    command ${binName} "$@"
    return
  fi

  local tmpfile
  tmpfile=$(mktemp) || {
    command ${binName} "$@"
    return $?
  }

  MINHTHETUS_SHELL_PIPE=$tmpfile command ${binName} "$@"
  local ret=$?

  if [ -s "$tmpfile" ]; then
    source "$tmpfile"
  fi
  rm -f "$tmpfile"
  return $ret
}
`;
    fs.writeFileSync(wrapperPath, wrapperContent.trim() + '\n');

    // 2. Modify the user's rc file to source the newly created static script
    if (fs.existsSync(configFile)) {
      let content = fs.readFileSync(configFile, 'utf8');
      
      const beginMarker = `# begin ${binName} shell wrapper`;
      const endMarker = `# end ${binName} shell wrapper`;
      
      // Remove old wrapper block if it exists
      content = content.replace(new RegExp(`${beginMarker}[\\s\\S]*?${endMarker}\\n?`, 'g'), '');
      
      // Append the new sourcing block
      const block = `
${beginMarker}
[ -f "${wrapperPath}" ] && source "${wrapperPath}"
${endMarker}
`;
      
      fs.writeFileSync(configFile, content.trim() + '\n' + block);
      console.log(`✅ Shell wrapper successfully added to ${configFile}.`);

      if (process.env.MINHTHETUS_SHELL_PIPE) {
        emitShellCommand(`source "${configFile}"`);
      }
    } else {
      console.error(`Shell config file not found at ${configFile}`);
    }
  } catch (err) {
    console.error('Shell wrapper setup failed:', err.message);
  }
}

module.exports = { setupShellWrapper };
