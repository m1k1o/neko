// emotes: the names the sprite sheet knows, showing one, sending one
import { app } from './app'
import { client, isMuted } from './client'

export const EMOTES = [
  'anger',
  'bomb',
  'sleep',
  'explode',
  'sweat',
  'poo',
  'hundred',
  'alert',
  'punch',
  'wave',
  'okay',
  'thumbs-up',
  'clap',
  'prey',
  'celebrate',
  'flame',
  'goof',
  'love',
  'cool',
  'smerk',
  'worry',
  'ouch',
  'cry',
  'surprised',
  'quiet',
  'rage',
  'annoy',
  'steamed',
  'scared',
  'terrified',
  'sleepy',
  'dead',
  'happy',
  'roll-eyes',
  'thinking',
  'clown',
  'sick',
  'rofl',
  'drule',
  'sniff',
  'sus',
  'party',
  'odd',
  'hot',
  'cold',
  'blush',
  'sad',
]

const s = app.state

export function showEmote(emote: string) {
  if (s.settings.ignore_emotes || document.visibilityState === 'hidden') return
  s.emotes[Math.random().toString(36).slice(2)] = emote
}

export function sendEmote(emote: string) {
  if (isMuted()) return
  client.sendBroadcast('emote', emote)
  showEmote(emote) // server does not echo broadcasts to the sender
}
