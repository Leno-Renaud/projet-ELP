import Player from "./player.js";
import Game from "./game.js";
import { ask, closePrompt } from "./prompt.js";

async function askInt(label, { min = 1, max = 100 } = {}) {
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const raw = await ask(label);
    const n = Number(raw);
    if (Number.isInteger(n) && n >= min && n <= max) return n;
    console.log(`Entrée invalide. Attendu: entier entre ${min} et ${max}.`);
  }
}

async function askNonEmpty(label) {
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const s = await ask(label);
    if (s.length > 0) return s;
    console.log("Entrée invalide. Nom non vide attendu.");
  }
}

async function main() {
  const n = await askInt("Nombre de joueurs (2-8) : ", { min: 2, max: 8 });

  const players = [];

  for (let i = 0; i < n; i++) {
    const name = await askNonEmpty(`Nom joueur ${i + 1}: `);
    players.push(new Player(name));
  }

  const game = new Game(players);

  await game.start();

  closePrompt();
}

main();
