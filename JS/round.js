import { CARD_TYPES } from "./card.js";
import { ask } from "./prompt.js";

export default class Round {
  constructor(deck, players, logger) {
    this.deck = deck;
    this.players = players;
    this.logger = logger;
    this.finished = false;
    this.flip7Player = null;
  }

  potentialScoreNow(player) {
    if (!player || player.frozen) return 0;

    let sum = Array.isArray(player.numbers)
      ? player.numbers.reduce((a, b) => a + b, 0)
      : 0;

    if (player.multiplier) sum *= 2;

    const bonus = Array.isArray(player.bonuses)
      ? player.bonuses.reduce((a, b) => a + b, 0)
      : 0;

    if (Array.isArray(player.numbers) && player.numbers.length === 7) sum += 15;

    return sum + bonus;
  }

  formatPlayerCards(player) {
    const cards = Array.isArray(player?.drawnCards) ? player.drawnCards : [];

    const numbers = cards
      .filter(c => c?.type === CARD_TYPES.NUMBER)
      .map(c => c.value)
      .filter(v => typeof v === "number")
      .sort((a, b) => a - b);

    const bonuses = cards
      .filter(c => c?.type === CARD_TYPES.BONUS)
      .map(c => c.value)
      .filter(v => typeof v === "number")
      .sort((a, b) => a - b);

    const effectLabels = {
      [CARD_TYPES.FREEZE]: "freeze",
      [CARD_TYPES.FLIP_THREE]: "flip three",
      [CARD_TYPES.SECOND_CHANCE]: "second chance",
      [CARD_TYPES.MULTIPLIER]: "x2"
    };

    const effects = cards
      .filter(c => c && c.type !== CARD_TYPES.NUMBER && c.type !== CARD_TYPES.BONUS)
      .map(c => effectLabels[c.type] ?? String(c.type))
      .filter(Boolean);

    const numbersLine = `Nombres: ${numbers.length ? numbers.join(", ") : "—"}`;
    const bonusLine = `Bonus: ${bonuses.length ? bonuses.map(v => `+${v}`).join(", ") : "—"}`;
    const effectsLine = `Effets: ${effects.length ? [...new Set(effects)].join(", ") : "—"}`;

    return [numbersLine, bonusLine, effectsLine].join("\n");
  }

  normalizeAction(raw) {
    const v = String(raw ?? "").trim().toLowerCase();
    if (["s", "stop", "rester", "stay"].includes(v)) return "s";
    if (["p", "piocher", "pioche", "draw", "d", "f", "flip"].includes(v)) return "p";
    return null;
  }

  async chooseTarget({ fromPlayer, effectLabel, allowSelf = true } = {}) {
    const eligible = allowSelf ? this.players : this.players.filter(p => p !== fromPlayer);

    if (eligible.length === 1) return eligible[0];

    console.log(`Choisir une cible pour ${effectLabel}:`);
    eligible.forEach((p, i) => {
      const flags = [
        p.active ? "actif" : "inactif",
        p.stayed ? "resté" : "",
        p.frozen ? "freeze" : ""
      ].filter(Boolean).join(", ");
      console.log(`  ${i + 1}) ${p.name}${flags ? ` (${flags})` : ""}`);
    });

    while (true) {
      const raw = await ask("Cible (numéro ou nom) : ");
      const byNum = Number(raw);
      if (Number.isInteger(byNum) && byNum >= 1 && byNum <= eligible.length) return eligible[byNum - 1];

      const byName = eligible.find(p => p.name.toLowerCase() === String(raw).toLowerCase());
      if (byName) return byName;

      console.log("Entrée invalide.");
    }
  }

  async chooseTargetFromList(targets, label) {
    if (!Array.isArray(targets) || targets.length === 0) return null;
    if (targets.length === 1) return targets[0];

    console.log(`Choisir une cible pour ${label}:`);
    targets.forEach((p, i) => console.log(`  ${i + 1}) ${p.name}`));




















    
    while (true) {
      const raw = await ask("Cible (numéro ou nom) : ");
      const byNum = Number(raw);
      if (Number.isInteger(byNum) && byNum >= 1 && byNum <= targets.length) return targets[byNum - 1];

      const byName = targets.find(p => p.name.toLowerCase() === String(raw).toLowerCase());
      if (byName) return byName;

      console.log("Entrée invalide.");
    }
  }

  async play(roundNumber) {
    this.roundOver = false;
    this.flip7Player = null;

    this.players.forEach(p => {
      p.resetRound();
      p.drawnCards = [];
    });

    this.logger.startRound(roundNumber, this.players);
    this.finished = false;
    this.flip7Player = null;

    for (const p of this.players) {
      await this.drawCard(p);
      if (this.roundOver) break;
    }

    while (!this.finished) {
      for (const p of this.players.filter(x => x.active && !x.stayed)) {
        if (this.roundOver) break;

        const potential = this.potentialScoreNow(p);
        console.log(`\n${p.name}: ${p.totalScore} points (${potential} points potentiels)`);
        console.log(this.formatPlayerCards(p));

        let choice = null;
        while (!choice) {
          const choiceRaw = await ask(`${p.name} → (p)iocher ou (s)rester ? `);
          choice = this.normalizeAction(choiceRaw);
          if (!choice) console.log("Entrée invalide (utiliser p/s)." );
        }

        if (choice === "s") {
          p.stayed = true;
          this.logger.log({ type: "stay", player: p.name });
          continue;
        }

        await this.drawCard(p);
        if (this.finished) break;
      }

      if (this.finished) break;

      const activeLeft = this.players.some(p => p.active && !p.stayed);
      if (!activeLeft) break;
    }

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
            const last = player.drawnCards[player.drawnCards.length - 1];
            if (last && last.type === CARD_TYPES.NUMBER && last.value === card.value) player.drawnCards.pop();
            const idx = [...player.drawnCards].reverse().findIndex(c => c.type === CARD_TYPES.SECOND_CHANCE);
            if (idx !== -1) {
              const realIndex = player.drawnCards.length - 1 - idx;
              player.drawnCards.splice(realIndex, 1);
            }

            this.logger.log({ type: "second_chance_used", player: player.name, duplicate: card.value });
            console.log(`${player.name}: second chance utilisée (doublon ${card.value} défaussé).`);
            break;
          }
          console.log("Doublon → éliminé !");
          player.active = false;
        } else {
          player.addNumber(card.value);

          if (player.numbers.length === 7) {
            console.log(`\n🎯 ${player.name} a fait FLIP 7 !`);
            console.log("Tous les joueurs actifs marquent leurs points !");
            
            for (const p of this.players) {
              if (p.active) {
                p.stayed = true;
              }
            }
            
            this.finished = true;
            this.flip7Player = player;
            this.logger.log({ type: "flip7", player: player.name });
          }
        }
        break;
      }

      case CARD_TYPES.FREEZE: {
        const frozenPlayer = await this.chooseTarget({
          fromPlayer: player,
          effectLabel: "FREEZE (finir le tour)",
          allowSelf: true
        });
        
        frozenPlayer.stayed = true;
        console.log(`${frozenPlayer.name} finit son tour et marque ses points !`);
        this.logger.log({ type: "freeze", player: player.name, targetPlayer: frozenPlayer.name });
        break;
      }

      case CARD_TYPES.FLIP_THREE: {
        const target = await this.chooseTarget({
          fromPlayer: player,
          effectLabel: "FLIP_THREE (3 cartes)",
          allowSelf: true
        });
        
        console.log(`${target.name} va recevoir 3 cartes...`);
        for (let i = 0; i < 3; i++) {
          if (this.finished) break;
          await this.drawCard(target);
        }
        break;
      }

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