// Set the real parent process timezone before Jest creates workers/VM contexts.
// Changing process.env.TZ inside a test's sandbox does not reliably change Date.
module.exports = async () => {
  process.env.TZ = 'Europe/London';
};
