export const logout = 'logg ut'
export const unsupported = 'Denne nettleseren støtter ikke WebRTC'
export const admin_loggedin = 'Du er innlogget som administrator'
export const you = 'Deg'
// TODO
//export const somebody = 'Somebody'
export const send_a_message = 'Send en melding'

export const side = {
  chat: 'Sludring',
  files: 'Filer',
  settings: 'Innstillinger',
}

export const connect = {
  login_title: 'Logg inn',
  invitation_title: 'Du har blitt invitert til dette rommet',
  displayname: 'Skriv inn ditt visningsnavn',
  password: 'Passord',
  connect: 'Koble til',
  error: 'Innloggingsfeil',
  // TODO
  //empty_displayname: 'Display Name cannot be empty.',
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
  // TODO
  //has: 'You have control',
  //hasnot: 'You do not have control',
}

export const locks = {
  // TODO
  //control: {
  //  lock: 'Lock Controls (for users)',
  //  unlock: 'Unlock Controls (for users)',
  //  locked: 'Controls Locked (for users)',
  //  unlocked: 'Controls Unlocked (for users)',
  //  notif_locked: 'locked controls for users',
  //  notif_unlocked: 'unlocked controls for users',
  //},
  login: {
    lock: 'Lås rommet (for brukere)',
    unlock: 'Lås opp rommet (for brukere)',
    locked: 'Rom låst (for brukere)',
    unlocked: 'Rom opplåst (for brukere)',
    notif_locked: 'låste rommet',
    notif_unlocked: 'låste opp rommet',
  },
  // TODO
  //file_transfer: {
  //  lock: 'Lock File Transfer (for users)',
  //  unlock: 'Unlock File Transfer (for users)',
  //  locked: 'File Transfer Locked (for users)',
  //  unlocked: 'File Transfer Unlocked (for users)',
  //  notif_locked: 'locked file transfer',
  //  notif_unlocked: 'unlocked file transfer',
  //},
}

export const setting = {
  scroll: 'Rullingssensitivitet',
  scroll_invert: 'Inverter rulling',
  autoplay: 'Spill video automatisk',
  ignore_emotes: 'Ignorer smilefjes',
  chat_sound: 'Sludringslyd',
  keyboard_layout: 'Tastaturoppsett',
  // TODO
  //broadcast_title: 'Live Broadcast',
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
}

export const files = {
  downloads: 'Overførsler',
  uploads: 'Overfør',
  upload_here: 'Klik eller træk filer her for at uploade',
}
