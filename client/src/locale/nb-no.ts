export const logout = 'logg ut'
export const unsupported = 'Denne nettleseren støtter ikke WebRTC'
export const admin_loggedin = 'Du er innlogget som administrator'
export const you = 'Deg'
export const send_a_message = 'Send en melding'

export const side = {
  chat: 'Sludring',
  settings: 'Innstillinger',
}

export const connect = {
  login_title: 'Logg inn',
  invitation_title: 'Du har blitt invitert til dette rommet',
  displayname: 'Skriv inn ditt visningsnavn',
  password: 'Passord',
  connect: 'Koble til',
  error: 'Innloggingsfeil',
}

export const context = {
  ignore: 'Ignorer',
  unignore: 'Opphev ignorering',
  mute: 'Forstum',
  unmute: 'Opphev forstummelse',
  release: 'Slipp kontrollen med tvang',
  take: 'Ta kontrollen med tvang',
  give: 'Gi vekk kontroll',
  kick: 'Kast ut',
  ban: 'Bannlys IP',
  confirm: {
    kick_title: 'Kast ut {name}?',
    kick_text: 'Vil du kaste ut {name}?',
    ban_title: 'Bannlys {name}?',
    ban_text: 'Vil du bannlyse {name}? Du vil måtte starte tjeneren på ny for å omgjøre dette.',
    mute_title: 'Mute {name}?',
    mute_text: 'Vil du forstumme {name}?',
    unmute_title: 'Unmute {name}?',
    unmute_text: 'Vil du oppheve forstummelsen av {name}?',
    button_yes: 'Ja',
    button_cancel: 'Avbryt',
  },
}

export const controls = {
  release: 'Slipp kontrollen',
  request: 'Forespør kontroll',
  lock: 'Lås kontrollen',
  unlock: 'Lås opp kontrollen',
}

export const room = {
  lock: 'Lås rommet (for brukere)',
  unlock: 'Lås opp rommet (for brukere)',
  locked: 'Rom låst (for brukere)',
  unlocked: 'Rom opplåst (for brukere)',
}

export const setting = {
  scroll: 'Rullingssensitivitet',
  scroll_invert: 'Inverter rulling',
  autoplay: 'Spill video automatisk',
  ignore_emotes: 'Ignorer smilefjes',
  chat_sound: 'Sludringslyd',
  keyboard_layout: 'Tastaturoppsett',
  broadcast_is_active: 'Kringkasting påslått',
  broadcast_url: 'RTMP-nettadresse',
}

export const connection = {
  logged_out: 'Du har blitt utlogget.',
  // TODO
  //reconnecting: 'Reconnecting',
  connected: 'Tilkoblet',
  disconnected: 'Frakoblet',
  // TODO
  //kicked: 'You have been removed from this room.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} koblet til',
  disconnected: '{name} koblet fra',
  controls_taken: '{name} tok kontrollen',
  controls_taken_force: 'tok kontrollen med tvang',
  controls_taken_steal: 'tok kontrollen fra {name}',
  controls_released: '{name} ga vekk kontrollen',
  controls_released_force: 'ga vekk kontrollen med tvang',
  controls_released_steal: 'ga vekk kontrollen fra {name}',
  controls_given: 'ga {name} kontrollen',
  controls_has: '{name} har kontrollen',
  controls_has_alt: 'Jeg la det komme vedkommende til kjenne at du ønsket den',
  controls_requesting: '{name} forespør kontrollen',
  resolution: 'endret oppløsningen til {width}x{height}@{rate}',
  banned: 'bannlyste {name}',
  kicked: 'kastet ut {name}',
  muted: 'forstummet {name}',
  unmuted: 'opphevet forstummingen av {name}',
  room_locked: 'låste rommet',
  room_unlocked: 'låste opp rommet',
}
