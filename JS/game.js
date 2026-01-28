import Deck from "./deck.js";
import Round from "./round.js";
import Logger from "./logger.js";

export default class Game {
  constructor(players) {
    this.players = players;
    this.logger = new Logger();
  }

  async start() {
    if (!Array.isArray(this.players) || this.players.length < 2) {
      throw new Error("Game.start: nombre de joueurs invalide (min 2)." );
    }

    let roundNumber = 1;

    while (!this.players.some(p => p.totalScore >= 200)) {
      console.log(`\n===== Manche ${roundNumber} =====`);

      const deck = new Deck();
      const round = new Round(deck, this.players, this.logger);

      await round.play(roundNumber);
      roundNumber++;
    }

    const winner = this.players.sort((a, b) => b.totalScore - a.totalScore)[0];

    console.log(`\n🏆 Gagnant : ${winner.name}`);
  }
}
