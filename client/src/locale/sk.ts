export const logout = 'odhlásiť sa'
export const unsupported = 'tento prehliadač nepodporuje webrtc'
export const admin_loggedin = 'Ste prihlásení ako administrátor'
export const you = 'Vy'// TODO: Incorrect in some translations!
export const send_a_message = 'Odoslať správu'

export const side = {
  chat: 'Chat',
  settings: 'Nastavenia',
}

export const connect = {
  title: 'Prosím prihláste sa',
  displayname: 'Zobrazované meno',
  password: 'Heslo',
  connect: 'Pripojiť sa',
}

export const context = {
  ignore: 'Ignorovať',
  unignore: 'Neignorovať',
  mute: 'Stlmiť zvuk',
  unmute: 'Zrušiť stlmenie zvuk',
  release: 'Vynutiť uvoľnenie ovládania',
  take: 'Prevziať ovládanie',
  give: 'Ponúknuť ovládanie',
  kick: 'Kick',
  ban: 'Ban IP',
  confirm: {
    kick_title: 'Kick {name}?',
    kick_text: 'Ste si istý, že chcete vykopnúť používateľa {name}?',
    ban_title: 'Ban {name}?',
    ban_text: 'Ste si istý, že chcete zablokovať používateľa {name}? Pre odblokovanie budete musieť reštartovať server.',
    mute_title: 'Stíšiť používateľa {name}?',
    mute_text: 'Ste si istý, že chcete stíšiť {name}?',
    unmute_title: 'Obnoviť zvuk používateľa {name}?',
    unmute_text: 'Ste si istý, že chcete obnoviť zvuk používateľa {name}?',
    button_yes: 'Áno',
    button_cancel: 'Zrušiť',
  }
}

export const controls = {
  release: 'Uvoľniť ovládanie',
  request: 'Požiadať o ovládanie',
  lock: 'Zamknúť ovládanie',
  unlock: 'Odomknúť ovládanie',
}

export const room = {
  lock: 'Zamknúť miestnosť',
  unlock: 'Odomknúť miestnosť',
  locked: 'Miestnosť zamknutá',
  unlocked: 'Miestnosť odomknutá',
}

export const setting = {
  scroll: 'Citlivosť kolieska myši',
  scroll_invert: 'Invertovať koliesko myši',
  autoplay: 'Automatické prehrávanie videa',
  ignore_emotes: 'Ignorovať smajlíky',
  chat_sound: 'Prehrávať zvuky chatu',
}

export const connection = {
  logged_out: 'Boli ste odhlásený',
  connected: 'Úspešne pripojený',
  disconnected: 'Boli ste odpojený',
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
  muted: '{name} bol/a stlmený/á',
  unmuted: '{name} už nie je viac stlmený/á',
  room_locked: 'miestnosť bola zamknutá',
  room_unlocked: 'miestnosť bola odomknutá',
}
