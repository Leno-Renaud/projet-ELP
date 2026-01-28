import { CARD_TYPES } from "./card.js";
import { ask } from "./prompt.js";

export default class Round {
  constructor(deck, players, logger) {
    this.deck = deck;
    this.players = players;
    this.logger = logger;
    this.finished = false;
    this.flip7Winner = null;
  }

  normalizeAction(raw) {
    const v = String(raw ?? "").trim().toLowerCase();
    if (["s", "stop", "rester", "stay"].includes(v)) return "s";
    if (["p", "piocher", "pioche", "draw", "d", "f", "flip"].includes(v)) return "p";
    return null;
  }

  async chooseTarget({ fromPlayer, effectLabel, allowSelf = true } = {}) {
    const eligible = allowSelf ? this.players : this.players.filter(p => p !== fromPlayer);

    // Si on ne peut pas choisir (ex: 1 seul joueur), fallback.
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

    // eslint-disable-next-line no-constant-condition
    while (true) {
      const raw = await ask("Cible (numéro ou nom) : ");
      const byNum = Number(raw);
      if (Number.isInteger(byNum) && byNum >= 1 && byNum <= eligible.length) return eligible[byNum - 1];

      const byName = eligible.find(p => p.name.toLowerCase() === String(raw).toLowerCase());
      if (byName) return byName;

      console.log("Entrée invalide.");
    }
  }

  async play(roundNumber) {
    this.players.forEach(p => {
      p.resetRound();
      // Historique des cartes piochées ce round
      p.drawnCards = [];
    });
    this.logger.startRound(roundNumber, this.players);
    this.finished = false;
    this.flip7Winner = null;

    // distribution initiale
    for (const p of this.players) {
      await this.drawCard(p);
    }

    while (!this.finished) {
      for (const p of this.players.filter(x => x.active && !x.stayed)) {
        console.log(`\n${p.name} a pioché:`, p.drawnCards.map(c => this.formatCard(c)).join(", "));

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

    // scoring
    for (const p of this.players) {
      const score = p.scoreRound();
      p.totalScore += score;
      console.log(`${p.name} gagne ${score} (total ${p.totalScore})`);
    }

    this.logger.endRound(this.players);
  }

  formatCard(card) {
    if (!card) return "Aucune carte";
    return card.value !== undefined ? `${card.type}(${card.value})` : `${card.type}`;
  }

  async drawCard(player) {
    const card = this.deck.draw();
    this.logger.log({ type: "draw", player: player.name, card });

    if (!card) return;

    // Stocke la carte dans l'historique du joueur (cartes "gardées" pour l'affichage)
    if (!player.drawnCards) player.drawnCards = [];
    player.drawnCards.push(card);

    // Affiche la carte tirée dans le terminal
    console.log(`${player.name} pioche: ${this.formatCard(card)}`);

    switch (card.type) {
      case CARD_TYPES.NUMBER:
        if (player.hasDuplicate(card.value)) {
          if (player.secondChance) {
            player.secondChance = false;
            // On défausse le doublon + on consomme la carte SECOND_CHANCE (retirée de l'affichage)
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
            player.stayed = true; // encaissement automatique
            this.finished = true;
            this.flip7Winner = player;
            this.logger.log({ type: "flip7", player: player.name });
            console.log(`${player.name} a fait FLIP 7 !`);
          }
        }
        break;

      case CARD_TYPES.FREEZE:
        // Carte à distribuer: on la "joue" et on la retire de l'affichage du joueur qui l'a piochée.
        player.drawnCards.pop();
        {
          const target = await this.chooseTarget({ fromPlayer: player, effectLabel: "FREEZE", allowSelf: false });
          this.logger.log({ type: "effect", effect: "freeze", from: player.name, to: target.name });
          console.log(`${player.name} joue FREEZE sur ${target.name}`);
          target.frozen = true;
          target.active = false;
        }
        break;

      case CARD_TYPES.FLIP_THREE:
        // Carte à distribuer: on la "joue" sur un joueur cible.
        player.drawnCards.pop();
        {
          const target = await this.chooseTarget({ fromPlayer: player, effectLabel: "FLIP3", allowSelf: false });
          this.logger.log({ type: "effect", effect: "flip_three", from: player.name, to: target.name });
          console.log(`${player.name} joue FLIP3 sur ${target.name}`);
          for (let i = 0; i < 3; i++) {
            if (!target.active || this.finished) break;
            await this.drawCard(target);
          }
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