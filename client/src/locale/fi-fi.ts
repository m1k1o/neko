export const logout = 'kirjaudu ulos'
export const unsupported = 'Tämä nettiselain ei tue WebRTC:tä'
export const admin_loggedin = 'Sinä olet kirjautunut valvojana'
export const you = 'Sinä'
export const somebody = 'Joku'
export const send_a_message = 'Lähetä viesti'

export const side = {
  chat: 'Chatti',
  settings: 'Asetukset',
}

export const connect = {
  login_title: 'Kirjaudu Sisään',
  invitation_title: 'Sinut on kutsuttu tähän huoneeseen',
  displayname: 'Kirjoita sinun näyttönimesi',
  password: 'Salasana',
  connect: 'Liity',
  error: 'Kirjautumis virhe',
  empty_displayname: 'Näyttönimesi ei voi olla tyhjä.',
}

export const context = {
  ignore: 'Estä',
  unignore: 'Poista esto',
  mute: 'Mykistä',
  unmute: 'Poista mykistys',
  release: 'Pakko vapauta kontrollit',
  take: 'Pakko ota kontrollit',
  give: 'Anna kontrollit',
  kick: 'Heitä ulos',
  ban: 'Kiellä IP',
  confirm: {
    kick_title: 'Haluatko heittää {name} ulos?',
    kick_text: 'Oletko varma että haluat heittää {name} ulos?',
    ban_title: 'Haluatko kieltää {name}?',
    ban_text: 'Haluatko kieltää {name}? Sinun pitää käynnistää palvelin uudestaan jos haluat kumota tämän.',
    mute_title: 'Haluatko mykistää {name}?',
    mute_text: 'Oletko varma että haluat mykistää {name}?',
    unmute_title: 'Poista {name} mykistys?',
    unmute_text: 'Oletko varma että haluat poistaa {name} mykistyksen?',
    button_yes: 'Kyllä',
    button_cancel: 'Peruuta',
  },
}

export const controls = {
  release: 'Vapauta kontrollit',
  request: 'Pyydä kontrollit',
  lock: 'Lukitse kontrollit',
  unlock: 'Vapauta kontrollit',
  has: 'Sinulla on kontrollit',
  hasnot: 'Sinulle ei ole kontrolleja',
}

export const locks = {
  control: {
    lock: 'Lock Controls (for users)',
    unlock: 'Unlock Controls (for users)',
    locked: 'Controls Locked (for users)',
    unlocked: 'Controls Unlocked (for users)',
    notif_locked: 'locked controls for users',
    notif_unlocked: 'unlocked controls for users',
  },
  login: {
    lock: 'Lock Room (for users)',
    unlock: 'Unlock Room (for users)',
    locked: 'Room Locked (for users)',
    unlocked: 'Room Unlocked (for users)',
    notif_locked: 'locked the room',
    notif_unlocked: 'unlocked the room',
  },
}

export const setting = {
  scroll: 'Scroll Sensitivity',
  scroll_invert: 'Invert Scroll',
  autoplay: 'Autoplay Video',
  ignore_emotes: 'Ignore Emotes',
  chat_sound: 'Play Chat Sound',
  keyboard_layout: 'Keyboard Layout',
  broadcast_title: 'Live Broadcast',
}

export const connection = {
  logged_out: 'You have been logged out.',
  reconnecting: 'Reconnecting...',
  connected: 'Connected',
  disconnected: 'Disconnected',
  kicked: 'You have been removed from this room.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} connected',
  disconnected: '{name} disconnected',
  controls_taken: '{name} took the controls',
  controls_taken_force: 'took the controls forcibly',
  controls_taken_steal: 'took the controls from {name}',
  controls_released: '{name} released the controls',
  controls_released_force: 'released the controls forcibly',
  controls_released_steal: 'released the controls from {name}',
  controls_given: 'gave the controls to {name}',
  controls_has: '{name} has the controls',
  controls_has_alt: 'But I let the person know you wanted it',
  controls_requesting: '{name} is requesting the controls',
  resolution: 'changed the resolution to {width}x{height}@{rate}',
  banned: 'banned {name}',
  kicked: 'kicked {name}',
  muted: 'muted {name}',
  unmuted: 'unmuted {name}',
}
