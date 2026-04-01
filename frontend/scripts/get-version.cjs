const { execSync } = require('child_process');
const { readFileSync } = require('fs');
const { join } = require('path');

/**
 * Get the application version based on git or package.json
 * Priority:
 * 1. Git tag for current commit (clean version)
 * 2. package.json version (fallback)
 *
 * Suffix rules:
 * - Exact tag match (release): no suffix
 * - main/master branch: no suffix
 * - Other branches: adds '-dev' suffix
 */
function getAppVersion() {
  let appVersion = '0.0.0-unknown';

  // 1. Try to get version from Git first
  try {
    const gitVersion = execSync('git describe --tags --always').toString().trim();
    // Remove everything after the first dash (commit count and hash)
    const cleanVersion = gitVersion.replace(/-.*$/, '');
    // Remove leading 'v' if present
    const version = cleanVersion.replace(/^v/, '');

    appVersion = version;

    // Check if we're on an exact tag (release build)
    // This is the most reliable way to detect tag-based builds
    let isRelease = false;
    try {
      execSync('git describe --tags --exact-match', { stdio: 'pipe' });
      isRelease = true;
    } catch {
      // Not on an exact tag
    }

    if (!isRelease) {
      // Check if we're on a feature branch (not main)
      // Add -dev suffix for non-main branches to indicate development build
      try {
        const currentBranch = execSync('git branch --show-current').toString().trim();
        // If currentBranch is empty (detached HEAD, like when a tag is checked out),
        // we assume it's a release and skip adding -dev.
        if (currentBranch && currentBranch !== 'main' && currentBranch !== 'master') {
          appVersion += '-dev';
        }
      } catch {
        // Couldn't determine branch, keep as-is
      }
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
    console.log('Using package.json version:', appVersion);
  } catch (e) {
    console.log('Using fallback version:', appVersion);
  }

  return appVersion;
}

module.exports = { getAppVersion };

// If called directly from command line
if (require.main === module) {
  console.log(getAppVersion());
}
