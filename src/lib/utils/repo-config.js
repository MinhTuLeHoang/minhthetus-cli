const fs = require('fs');
const path = require('path');
const { getSystemInfo } = require('./system-info');

const BIN_NAME = 'minhthetus-cli';

/**
 * Adds a repository to the ~/.minhthetus-cli/list-repo.json file.
 * 
 * @param {string} repoPath - The absolute path to the repository.
 * @param {Object} options - Configuration options.
 */
function addRepo(repoPath, options = {}) {
  const silent = options.silent || false;
  const { homeDir } = getSystemInfo();
  const configDir = path.join(homeDir, `.${BIN_NAME}`);
  const configFile = path.join(configDir, 'list-repo.json');

  if (!fs.existsSync(configDir)) {
    fs.mkdirSync(configDir, { recursive: true });
  }

  let repos = [];
  if (fs.existsSync(configFile)) {
    try {
      repos = JSON.parse(fs.readFileSync(configFile, 'utf8'));
    } catch (e) {
      console.error('Error parsing list-repo.json, resetting to empty array.');
      repos = [];
    }
  }

  // Get info from package.json
  const packageJsonPath = path.join(repoPath, 'package.json');
  let name = path.basename(repoPath);
  let description = '';

  if (fs.existsSync(packageJsonPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
      name = pkg.name || name;
      description = pkg.description || '';
    } catch (e) {
      // Ignore parsing errors for package.json
    }
  }

  // Check if repo already exists in list (by path)
  const existingIndex = repos.findIndex(r => r.path === repoPath);
  const repoConfig = {
    name,
    description,
    path: repoPath
  };

  if (existingIndex > -1) {
    repos[existingIndex] = repoConfig;
  } else {
    repos.push(repoConfig);
  }

  fs.writeFileSync(configFile, JSON.stringify(repos, null, 2));
  if (!silent) {
    console.log(`✅ Repository '${name}' added to tracking list.`);
  }
}

/**
 * Removes a repository from the ~/.minhthetus-cli/list-repo.json file.
 * 
 * @param {string} repoPath - The absolute path to the repository.
 */
function removeRepo(repoPath) {
  const { homeDir } = getSystemInfo();
  const configDir = path.join(homeDir, `.${BIN_NAME}`);
  const configFile = path.join(configDir, 'list-repo.json');

  if (!fs.existsSync(configFile)) return;

  try {
    let repos = JSON.parse(fs.readFileSync(configFile, 'utf8'));
    const initialLength = repos.length;
    repos = repos.filter(r => r.path !== repoPath);
    
    if (repos.length < initialLength) {
      fs.writeFileSync(configFile, JSON.stringify(repos, null, 2));
      console.log(`🗑️ Repository at ${repoPath} removed from tracking list.`);
    }
  } catch (e) {
    console.error('Error updating list-repo.json:', e.message);
  }
}

module.exports = { addRepo, removeRepo };
