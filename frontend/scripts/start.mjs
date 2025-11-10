import { execSync } from 'child_process';
import dotenv from 'dotenv';
const nodeEnv = process.env.NODE_ENV || 'production';

const envFile = `.env.${nodeEnv}`;
console.log(`Loading env file: ${envFile}`);

const result = dotenv.config({ path: envFile });
if (result.error) {
  console.error(`Failed to load env file: ${envFile}`, result.error);
  process.exit(1);
}

const port = process.env.PORT || 9080; // default to vite preview port
console.log(`Using env file: ${envFile}`);
console.log(`Starting static server on port ${port}`);

execSync(`serve -s dist -l tcp://0.0.0.0:${port}`, { stdio: 'inherit' });
