import readline from "readline";

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

export const ask = (q) =>
  new Promise(resolve => rl.question(q, (answer) => resolve(String(answer).trim())));

export const closePrompt = () => rl.close();
