// emotes: the names the sprite sheet knows, showing one, sending one
import { isMuted, type AppStore, type NekoApp } from './app'
import { uid } from '@/lib/utils'

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

export function showEmote(app: AppStore, emote: string) {
  if (app.getState().settings.ignore_emotes || document.visibilityState === 'hidden') return
  app.setState((s) => ({ emotes: { ...s.emotes, [uid()]: emote } }))
}

// the emote's animation is over
export function hideEmote(app: AppStore, id: string) {
  app.setState((s) => {
    const emotes = { ...s.emotes }
    delete emotes[id]
    return { emotes }
  })
}

export function sendEmote({ client, app }: NekoApp, emote: string) {
  if (isMuted(client)) return
  client.sendBroadcast('emote', emote)
  showEmote(app, emote) // server does not echo broadcasts to the sender
}
