import { CARD_TYPES, numberCard } from "./card.js";

export default class Deck {
  constructor() {
    this.cards = [];
    this.build();
    this.shuffle();
  }

  build() {
    // nombres : 0 en 1 exemplaire, puis N exemplaires de N (1..12)
    this.cards.push(numberCard(0));
    for (let v = 1; v <= 12; v++) {
      for (let i = 0; i < v; i++) this.cards.push(numberCard(v));
    }

    // actions 
    for (let i = 0; i < 3; i++) this.cards.push({ type: CARD_TYPES.FREEZE });
    for (let i = 0; i < 3; i++) this.cards.push({ type: CARD_TYPES.FLIP_THREE });
    for (let i = 0; i < 3; i++) this.cards.push({ type: CARD_TYPES.SECOND_CHANCE });

    // bonus +2 à +10
    for (let v = 2; v <= 10; v++) {
      this.cards.push({ type: CARD_TYPES.BONUS, value: v });
    }

    // multiplier x2
    this.cards.push({ type: CARD_TYPES.MULTIPLIER });
  }

  shuffle() {
    for (let i = this.cards.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [this.cards[i], this.cards[j]] = [this.cards[j], this.cards[i]];
    }
  }

  draw() {
    return this.cards.pop();
  }
}
