export const logout = 'kirjaudu ulos'
export const unsupported = 'Tämä nettiselain ei tue WebRTC:tä'
export const admin_loggedin = 'Sinä olet kirjautunut valvojana'
export const you = 'Sinä'
export const somebody = 'Joku'
export const send_a_message = 'Lähetä viesti'

export const side = {
  chat: 'Chatti',
  files: 'Tiedostot',
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
    lock: 'Lukitse kontrollit (käyttäjiltä)',
    unlock: 'Vapauta kontrollit (käyttäjiltä)',
    locked: 'Kontrollit lukittu (käyttäjiltä)',
    unlocked: 'Kontrollit vapautettu (käyttäjiltä)',
    notif_locked: 'kontrollit on lukittu käyttäjiltä',
    notif_unlocked: 'kontrollit on vapautettu käyttäjille',
  },
  login: {
    lock: 'Lukitse huone (käyttäjiltä)',
    unlock: 'Vapauta huone (käyttäjiltä)',
    locked: 'Huone lukittu (käyttäjiltä)',
    unlocked: 'Huone vapautettu (käyttäjiltä)',
    notif_locked: 'lukittu huone',
    notif_unlocked: 'vapautettu huone',
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
  scroll: 'Scrollin herkkyys',
  scroll_invert: 'Käänteinen Scroll',
  autoplay: 'Automaattisesti toista video',
  ignore_emotes: 'Estä emojit',
  chat_sound: 'Soita viesti ääni',
  keyboard_layout: 'Näppäimistöasettelu',
  broadcast_title: 'Suora Lähetys',
}

export const connection = {
  logged_out: 'Sinut on kirjattu ulos.',
  reconnecting: 'Yhteyttä yritetään palauttaa...',
  connected: 'Yhdistetty',
  disconnected: 'Katkaistu yhteys',
  kicked: 'Sinut on poistettu huoneesta.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} liittyi',
  disconnected: '{name} poistui',
  controls_taken: '{name} otti kontrollit',
  controls_taken_force: 'otti kontrollit pakolla',
  controls_taken_steal: 'otti kontrollit {name}',
  controls_released: '{name} vapautti kontrollit',
  controls_released_force: 'vapautti kontrollit pakolla',
  controls_released_steal: 'vapautti kontrollit {name}',
  controls_given: 'antoi kontrollit {name}',
  controls_has: '{name} on kontrollit',
  controls_has_alt: 'Kerroin henkilölle että haluat ne',
  controls_requesting: '{name} pyytää kontrolleja',
  resolution: 'vaihdettu resoluutio {width}x{height}@{rate}',
  banned: 'kielletty {name}',
  kicked: 'heitetty {name} ulos',
  muted: 'mykistetty {name}',
  unmuted: 'poistettu mykistys {name}',
}

export const files = {
  downloads: 'Lataukset',
  uploads: 'Lataa',
  upload_here: 'Klikkaa tai vedä tiedostoja tähän ladataksesi',
}
