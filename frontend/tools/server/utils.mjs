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

export function createMockToken(payload, expiresInMinutes) {
  const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');

  const now = Math.floor(Date.now() / 1000); // current time in seconds
  const bodyPayload = { ...payload };

  if (expiresInMinutes) {
    bodyPayload.exp = now + expiresInMinutes * 60; 
  }

  const body = Buffer.from(JSON.stringify(bodyPayload)).toString('base64url');
  const signature = Buffer.from('mock_signature').toString('base64url');

  return `${header}.${body}.${signature}`;
}
