import { CARD_TYPES } from "./card.js";
import { ask } from "./prompt.js";

export default class Round {
  constructor(deck, players, logger) {
    this.deck = deck;
    this.players = players;
    this.logger = logger;

    // état du round
    this.roundOver = false;
    this.flip7Player = null;
  }

  async play(roundNumber) {
    // reset état round
    this.roundOver = false;
    this.flip7Player = null;

    this.players.forEach(p => {
      p.resetRound();
      p.drawnCards = [];
    });

    this.logger.startRound(roundNumber, this.players);

    // distribution initiale
    for (const p of this.players) {
      await this.drawCard(p);
      if (this.roundOver) break;
    }

    while (!this.roundOver) {
      for (const p of this.players.filter(x => x.active && !x.stayed)) {
        if (this.roundOver) break;

        console.log(`\n${p.name} a pioché:`, p.drawnCards.map(c => this.formatCard(c)).join(", "));
        const choice = await ask(`${p.name} → (p)iocher ou (s)rester ? `);

        if (choice === "s") {
          p.stayed = true;
          this.logger.log({ type: "stay", player: p.name });
          continue;
        }

        await this.drawCard(p);
        if (this.roundOver) break;
      }

      if (this.roundOver) break;

      const activeLeft = this.players.some(p => p.active && !p.stayed);
      if (!activeLeft) break;
    }

    if (this.flip7Player) {
      console.log(`${this.flip7Player.name} a fait FLIP 7 !`);
      this.logger.log({ type: "flip7", player: this.flip7Player.name });
    }

    // scoring UNE SEULE FOIS, fin de round
    this.scoreRound();

    this.logger.endRound(this.players);
  }

  formatCard(card) {
    if (!card) return "Aucune carte";
    return card.value !== undefined ? `${card.type}(${card.value})` : `${card.type}`;
  }

  scoreRound() {
    for (const p of this.players) {
      const score = p.scoreRound();
      p.totalScore += score;

      console.log(`${p.name} gagne ${score} (total ${p.totalScore})`);
      this.logger.log({ type: "roundScore", player: p.name, score, totalScore: p.totalScore });
    }
  }

  async drawCard(player) {
    if (this.roundOver) return;

    const card = this.deck.draw();
    this.logger.log({ type: "draw", player: player.name, card });

    if (!card) return;

    if (!player.drawnCards) player.drawnCards = [];
    player.drawnCards.push(card);

    console.log(`${player.name} pioche: ${this.formatCard(card)}`);

    switch (card.type) {
      case CARD_TYPES.NUMBER: {
        if (player.hasDuplicate(card.value)) {
          if (player.secondChance) {
            player.secondChance = false;
            console.log("Second chance utilisée !");
            break;
          }
          console.log("Doublon → éliminé !");
          player.active = false;
          break;
        }

        player.addNumber(card.value);

        // Déclenche FLIP7 immédiatement (important pour FLIP_THREE)
        if (player.numbers.length === 7) {
          this.roundOver = true;
          this.flip7Player = player;
        }
        break;
      }

      case CARD_TYPES.FREEZE: {
        const activePlayers = this.players.filter(p => p.active && p !== player);
        
        if (activePlayers.length === 0) {
          console.log("Aucun autre joueur actif !");
          break;
        }

        console.log("\nJoueurs actifs disponibles :");
        activePlayers.forEach((p, index) => {
          console.log(`${index + 1}. ${p.name}`);
        });

        let selectedIndex = -1;
        while (selectedIndex < 0 || selectedIndex >= activePlayers.length) {
          const input = await ask(`${player.name}, choisis un joueur (1-${activePlayers.length}) : `);
          selectedIndex = parseInt(input) - 1;
        }

        const frozenPlayer = activePlayers[selectedIndex];
        frozenPlayer.stayed = true;
        console.log(`${frozenPlayer.name} finit son tour et marque ses points !`);
        this.logger.log({ type: "freeze", player: player.name, targetPlayer: frozenPlayer.name });
        break;
      }

      case CARD_TYPES.FLIP_THREE:
        for (let i = 0; i < 3; i++) {
          if (this.roundOver) break;
          await this.drawCard(player);
        }
        break;

      case CARD_TYPES.SECOND_CHANCE:
        player.secondChance = true;
        break;

      case CARD_TYPES.BONUS:
        player.bonuses.push(card.value);
        break;

      case CARD_TYPES.MULTIPLIER:
        player.multiplier = true;
        break;
    }
  }
}