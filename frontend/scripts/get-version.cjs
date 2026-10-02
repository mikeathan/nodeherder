const { execFileSync } = require('child_process');
const { readFileSync } = require('fs');
const { join } = require('path');

/**
 * Get the application version based on git or package.json
 * Priority:
 * 1. Git tag for current commit (clean version)
 * 2. package.json version (fallback)
 *
 * Adds '-dev' suffix for non-main branches
 */
function getAppVersion() {
  let appVersion = '0.0.0-unknown';
  const gitOptions = { cwd: join(__dirname, '..', '..'), encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] };

  // 1. Try to get version from Git first
  try {
    const gitVersion = execFileSync('git', ['describe', '--tags', '--always'], gitOptions).trim();
    // Remove everything after the first dash (commit count and hash)
    const cleanVersion = gitVersion.replace(/-.*$/, '');
    // Remove leading 'v' if present
    const version = cleanVersion.replace(/^v/, '');

    appVersion = version;

    // Check if we're on a feature branch (not main)
    // Add -dev suffix for non-main branches to indicate development build
    try {
      const currentBranch = execFileSync('git', ['branch', '--show-current'], gitOptions).trim();
      // If currentBranch is empty (detached HEAD, like when a tag is checked out), 
      // we assume it's a release and skip adding -dev.
      if (currentBranch && currentBranch !== 'main' && currentBranch !== 'master') {
        appVersion += '-dev';
      }
    } catch {
      // Couldn't determine branch, keep as-is
    }

    return appVersion; // Got pure git version, return it
  } catch (error) {
    // Git is not available, proceed to fallback
  }

  // 2. Fallback to package.json
  try {
    const packageJsonPath = join(__dirname, '..', 'package.json');
    const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf8'));
    appVersion = packageJson.version;
  } catch (e) {
    // Keep stdout limited to the version for shell callers.
  }

  return appVersion;
}

module.exports = { getAppVersion };

// If called directly from command line
if (require.main === module) {
  console.log(getAppVersion());
}
