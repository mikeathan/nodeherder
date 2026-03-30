const { execSync } = require('child_process');
const { readFileSync } = require('fs');
const { join } = require('path');

/**
 * Get the application version based on git and package.json
 * Priority:
 * 1. Git tag for current commit (clean version)
 * 2. Git describe output (nearest tag)
 * 3. package.json version (fallback)
 *
 * Adds '-dev' suffix for non-main branches
 */
function getAppVersion() {
  const packageJsonPath = join(process.cwd(), 'package.json');
  const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf8'));
  let appVersion = packageJson.version;

  try {
    // Get git describe output
    const gitVersion = execSync('git describe --tags --always').toString().trim();

    // Remove everything after the first dash (commit count and hash)
    // v0.12.13-1-gfa05ab14 → v0.12.13
    // v0.12.13 → v0.12.13
    // fa05ab14 → fa05ab14 (no tags, just commit hash)
    const cleanVersion = gitVersion.replace(/-.*$/, '');

    // Remove leading 'v' if present
    appVersion = cleanVersion.replace(/^v/, '');

    // Check if we're on a feature branch (not main)
    // Add -dev suffix for non-main branches to indicate development build
    try {
      const currentBranch = execSync('git branch --show-current').toString().trim();
      if (currentBranch !== 'main' && currentBranch !== 'master') {
        appVersion = appVersion + '-dev';
      }
    } catch {
      // Couldn't determine branch, keep as-is
    }
  } catch (error) {
    // Fallback to package.json if git is not available
    console.log('Using package.json version:', appVersion);
  }

  return appVersion;
}

module.exports = { getAppVersion };

// If called directly from command line
if (require.main === module) {
  console.log(getAppVersion());
}
