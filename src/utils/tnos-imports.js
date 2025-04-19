/**
 * WHO: TNOSImportHelper
 * WHAT: Path resolution and import helper for TNOS architecture
 * WHEN: During module loading and initialization
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To ensure consistent path resolution across TNOS components
 * HOW: By providing standardized paths and import utilities
 * EXTENT: All system modules
 */

const path = require('path');
const fs = require('fs');

// Determine TNOS root path
const TNOS_ROOT = process.env.TNOS_ROOT || '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS';

// Create a context-aware path resolver
function getTNOSPaths() {
  return {
    root: TNOS_ROOT,
    core: {
      base: path.join(TNOS_ROOT, 'src/main/resources/core'),
      cognition: path.join(TNOS_ROOT, 'src/main/resources/core/cognition'),
      reactive: path.join(TNOS_ROOT, 'src/main/resources/reactive')
    },
    atl: path.join(TNOS_ROOT, 'atl'),
    agents: path.join(TNOS_ROOT, 'agents'),
    mcp: {
      base: path.join(TNOS_ROOT, 'mcp'),
      server: path.join(TNOS_ROOT, 'mcp/server'),
      integration: path.join(TNOS_ROOT, 'mcp/integration'),
      temp: path.join(TNOS_ROOT, 'mcp/temp')
    },
    logs: path.join(TNOS_ROOT, 'logs')
  };
}

// Ensure all critical directories exist
function ensureDirectories() {
  const paths = getTNOSPaths();
  const criticalPaths = [
    paths.core.base,
    paths.core.cognition,
    paths.core.reactive,
    paths.atl,
    paths.agents,
    paths.mcp.base,
    paths.mcp.server,
    paths.mcp.integration,
    paths.mcp.temp,
    paths.logs
  ];
  
  for (const dir of criticalPaths) {
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
      console.log(`Created directory: ${dir}`);
    }
  }
}

// Export functions
module.exports = {
  getPaths: getTNOSPaths,
  ensureDirectories
};

