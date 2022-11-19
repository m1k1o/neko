export const logout = 'odhlásiť sa'
export const unsupported = 'tento prehliadač nepodporuje webrtc'
export const admin_loggedin = 'Ste prihlásení/á ako administrátor'
// export const you = '' // Incorrect in some translations! Cannot be used!
// TODO
//export const somebody = 'Somebody'
export const send_a_message = 'Odoslať správu'

export const side = {
  chat: 'Chat',
  files: 'Súbory',
  settings: 'Nastavenia',
}

export const connect = {
  login_title: 'Prihláste sa',
  invitation_title: 'Boli ste pozvaný/á do miestnosti',
  displayname: 'Vaše meno',
  password: 'Heslo',
  connect: 'Pripojiť sa',
  error: 'Chyba pri prihlasovaní',
  empty_displayname: 'Meno nemôže byť prázdne.',
}

export const context = {
  ignore: 'Ignorovať',
  unignore: 'Zrušiť ignorovanie',
  mute: 'Zakázať chat',
  unmute: 'Povoliť chat',
  release: 'Zrušiť ovládanie',
  take: 'Prevziať ovládanie',
  give: 'Ponúknuť ovládanie',
  kick: 'Kick',
  ban: 'Ban IP',
  confirm: {
    kick_title: 'Kick {name}?',
    kick_text: 'Ste si istý/á, že chcete vykopnúť používateľa {name}?',
    ban_title: 'Ban {name}?',
    ban_text:
      'Ste si istý/á, že chcete zablokovať používateľa {name}? Pre odblokovanie budete musieť reštartovať server.',
    mute_title: 'Zakázať chat pre používateľa {name}?',
    mute_text: 'Ste si istý/á, že chcete zakázať chat pre používateľa {name}?',
    unmute_title: 'Povoliť chat pre používateľa {name}?',
    unmute_text: 'Ste si istý/á, že chcete povoliť chat pre používateľa {name}?',
    button_yes: 'Áno',
    button_cancel: 'Zrušiť',
  },
}

export const controls = {
  release: 'Uvoľniť ovládanie',
  request: 'Požiadať o ovládanie',
  lock: 'Zamknúť ovládanie',
  unlock: 'Odomknúť ovládanie',
  // TODO
  //has: 'You have control',
  //hasnot: 'You do not have control',
}

export const locks = {
  control: {
    lock: 'Zakázať ovládanie (pre používateľov)',
    unlock: 'Povoliť ovládanie (pre používateľov)',
    locked: 'Ovládanie je zakázané (pre používateľov)',
    unlocked: 'Ovládanie je povolené (pre používateľov)',
    notif_locked: 'zakázal/a ovládanie pre používateľov',
    notif_unlocked: 'povolil/a ovládanie pre používateľov',
  },
  login: {
    lock: 'Zamknúť miestnosť (pre používateľov)',
    unlock: 'Odomknúť miestnosť (pre používateľov)',
    locked: 'Miestnosť je zamknutá (pre používateľov)',
    unlocked: 'Miestnosť odomknutá (pre používateľov)',
    notif_locked: 'miestnosť bola zamknutá',
    notif_unlocked: 'miestnosť bola odomknutá',
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
  scroll: 'Citlivosť kolieska myši',
  scroll_invert: 'Invertovať koliesko myši',
  autoplay: 'Automatické prehrávanie videa',
  ignore_emotes: 'Ignorovať smajlíky',
  chat_sound: 'Prehrávať zvuky chatu',
  keyboard_layout: 'Rozloženie klávesnice',
  broadcast_title: 'Živé vysielanie',
}

export const connection = {
  logged_out: 'Boli ste odhlásený/á',
  reconnecting: 'Obnova spojenia...',
  connected: 'Úspešne pripojený/á',
  disconnected: 'Boli ste odpojený/á',
  kicked: 'Boli ste odstránený/á z tejto miestnosti.',
  button_confirm: 'Ok',
}

export const notifications = {
  connected: '{name} sa pripojil/a',
  disconnected: '{name} sa odpojil/a',
  controls_taken: '{name} prevzal/a ovládanie',
  controls_taken_force: 'ovládanie bolo prevzaté',
  controls_taken_steal: 'prevzal/a ovládanie od použivateľa {name}',
  controls_released: '{name} uvoľnil/a ovládanie',
  controls_released_force: 'ovládanie bolo uvoľnené',
  controls_released_steal: 'uvoľnil/a ovládanie použivateľa {name}',
  controls_given: 'ponúkol/a ovládanie používateľovi {name}',
  controls_has: '{name} má ovládanie',
  controls_has_alt: 'Ale dám mu vedieť, že si chcel ovládanie',
  controls_requesting: '{name} by chcel/a ovládanie',
  resolution: 'zmenené rozlíšenie na {width}x{height}@{rate}',
  banned: '{name} dostal/a BAN',
  kicked: '{name} bol/a vykopnutý/a',
  muted: 'zakázal chat používateľovi {name}',
  unmuted: 'povolil chat používateľovi {name}',
}

export const files = {
  downloads: 'Stiahnutia',
  uploads: 'Nahrávanie',
  upload_here: 'Kliknutím alebo pretiahnutím súborov sem ich môžete nahrať',
}
