const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');
const os = require('os');

const GUM_VERSION = '0.14.5';
const VENDOR_DIR = path.join(__dirname, '../../vendor');
const BIN_DIR = path.join(VENDOR_DIR, 'bin');
const GUM_PATH = path.join(BIN_DIR, 'gum');

async function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        downloadFile(response.headers.location, dest).then(resolve).catch(reject);
        return;
      }
      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: ${response.statusCode}`));
        return;
      }
      response.pipe(file);
      file.on('finish', () => {
        file.close(resolve);
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => reject(err));
    });
  });
}

function getPlatform() {
  const platform = os.platform();
  if (platform === 'darwin') return 'Darwin';
  if (platform === 'linux') return 'Linux';
  throw new Error(`Unsupported platform: ${platform}`);
}

function getArch() {
  const arch = os.arch();
  if (arch === 'x64') return 'x86_64';
  if (arch === 'arm64') return 'arm64';
  throw new Error(`Unsupported architecture: ${arch}`);
}

async function installGum() {
  if (fs.existsSync(GUM_PATH)) {
    console.log('Gum is already installed.');
    return;
  }

  if (!fs.existsSync(BIN_DIR)) {
    fs.mkdirSync(BIN_DIR, { recursive: true });
  }

  const platform = getPlatform();
  const arch = getArch();
  const filename = `gum_${GUM_VERSION}_${platform}_${arch}.tar.gz`;
  const url = `https://github.com/charmbracelet/gum/releases/download/v${GUM_VERSION}/${filename}`;
  const tempPath = path.join(os.tmpdir(), filename);

  console.log(`Downloading gum v${GUM_VERSION} for ${platform} ${arch}...`);
  await downloadFile(url, tempPath);

  console.log('Extracting gum...');
  try {
    // Extract only the gum binary from the tarball
    execSync(`tar -xzf "${tempPath}" -C "${BIN_DIR}" --strip-components=1`, { stdio: 'inherit' });
    // Some tarballs might have gum at root or in a subdir, let's check
    // If double check is needed:
    if (!fs.existsSync(GUM_PATH)) {
        // Fallback for different tar structures if any
        execSync(`tar -xzf "${tempPath}" -C "${BIN_DIR}"`, { stdio: 'inherit' });
        // find gum and move it to BIN_DIR/gum if it's somewhere else
    }
  } catch (err) {
    console.error('Failed to extract gum:', err);
    throw err;
  } finally {
    if (fs.existsSync(tempPath)) {
      fs.unlinkSync(tempPath);
    }
  }

  if (fs.existsSync(GUM_PATH)) {
    fs.chmodSync(GUM_PATH, 0o755);
    console.log('Gum installed successfully at', GUM_PATH);
  } else {
    throw new Error('Gum binary not found after extraction.');
  }
}

if (require.main === module) {
  installGum().catch((err) => {
    console.error(err);
    process.exit(1);
  });
}

module.exports = { installGum, GUM_PATH };
