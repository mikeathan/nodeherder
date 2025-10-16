/* eslint-disable no-console */
import moment from 'moment';
import 'moment-timezone';
import { createRequire } from 'module';
import { TIMEZONE } from './config.mjs';

export function currentTime() {
  return moment().tz(TIMEZONE).format();
}

export function loadJson(path) {
  const require = createRequire(import.meta.url);
  return require(path);
}

export function sendMessage(ws, event, payload) {
  const msg = JSON.stringify({ type: event, payload });
  ws.send(msg);
}

export function sendOperationSuccess(ws) {
  sendMessage(ws, 'operationSuccess', {});
}

export function sendOperationFailed(ws, message) {
  sendMessage(ws, 'operationFailed', message);
}

export function unpackJsonToMap(value) {
  try {
    return JSON.parse(value);
  } catch (e) {
    console.log('[ERROR] unpacking json', value, e);
    return null;
  }
}
