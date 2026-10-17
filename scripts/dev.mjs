#!/usr/bin/env node
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const children = new Set();
const groups = new Set();
let stopping = false;
let shutdownTimer;

function signalGroups(signal) {
  for (const pid of groups) {
    try {
      process.kill(-pid, signal);
    } catch (error) {
      if (error.code !== 'ESRCH') console.error(`Cannot send ${signal}: ${error.message}`);
    }
  }
}

function finishIfStopped() {
  if (!stopping || children.size) return;
  // npm or go can exit before their descendants; do not leave those running.
  signalGroups('SIGKILL');
  clearTimeout(shutdownTimer);
}

function stop(code) {
  if (stopping) return;
  stopping = true;
  process.exitCode = code;
  signalGroups('SIGTERM');
  shutdownTimer = setTimeout(() => signalGroups('SIGKILL'), 5000);
  finishIfStopped();
}

function run(label, command, args, directory, onSuccess) {
  if (stopping) return;
  console.log(`[dev] ${label}`);
  const child = spawn(command, args, {
    cwd: `${root}${directory}`,
    env: { ...process.env, APP_ENV: process.env.APP_ENV || 'development' },
    stdio: 'inherit',
    // Separate process groups let shutdown reach subprocesses as well as npm/go.
    detached: true,
  });
  children.add(child);
  if (child.pid) groups.add(child.pid);
  child.on('error', (error) => {
    console.error(`[dev] ${label}: ${error.message}`);
    children.delete(child);
    stop(1);
    finishIfStopped();
  });
  child.on('exit', (code, signal) => {
    children.delete(child);
    if (!stopping) {
      if (code === 0 && onSuccess) {
        groups.delete(child.pid);
        onSuccess();
      } else {
        console.error(`[dev] ${label} exited (${signal || code}); stopping services.`);
        stop(code || 1);
      }
    }
    finishIfStopped();
  });
}

process.on('SIGINT', () => stop(130));
process.on('SIGTERM', () => stop(143));

if (process.platform === 'win32') {
  console.error('[dev] This launcher requires macOS or Linux.');
  process.exitCode = 1;
} else {
  run('Building backend', 'go', ['build', '-o', 'nodeherder', '.'], 'backend', () => {
    run('Backend', './nodeherder', [], 'backend');
    run('Frontend', 'npm', ['run', 'dev', '--', '--strictPort'], 'frontend');
  });
}
