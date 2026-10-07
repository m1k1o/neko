export const logout = 'Ausloggen'
export const unsupported = 'Dieser Webbrowser unterstützt kein Web-RTC.'
export const admin_loggedin = 'Du bist eingeloggt als Admin.'
export const you = 'Du'
export const somebody = 'Jemand'

export const side = {
  settings: 'Einstellungen',
}

export const connect = {
  login_title: 'Bitte Anmelden',
  invitation_title: 'Du wurdest zu diesem Raum eingeladen',
  displayname: 'Gebe deinen Benutzernamen an',
  password: 'Passwort',
  connect: 'Verbinden',
  error: 'Login Fehler',
  empty_displayname: 'Benutzername kann nicht leer sein.',
}

export const context = {
  ignore: 'Ignorieren',
  unignore: 'Nicht Ignorieren',
  release: 'Freigabesteuerung freigeben',
  take: 'Steuerung erzwingen',
  give: 'Steuerung geben',
  kick: 'Rauswerfen',
  confirm: {
    kick_title: 'Kick {name}?',
    kick_text: 'Bist du sicher das du {name} rauswerfen willst?',
    ban_title: '{name} Sperren?',
    button_yes: 'Ja',
    button_cancel: 'Abbrechen',
  },
}

export const controls = {
  release: 'Steuerung freigeben',
  request: 'Steuerung anfordern',
  lock: 'Steuerung sperren',
  unlock: 'Steuerung entsperren',
  has: 'Du hast die Kontrolle', // TODO: prüfen Übersetzung
  hasnot: 'Du hast keine Kontrolle', // TODO: prüfen Übersetzung
}

export const locks = {
  control: {
    lock: 'Steuerung sperren (für Nutzer)',
    unlock: 'Steuerung entsperren (für Nutzer)',
    locked: 'Steuerung gesperrt (für Nutzer)',
    unlocked: 'Steuerung entsperrt (für Nutzer)',
    notif_locked: 'Steuerung sperren für Nutzer',
    notif_unlocked: 'Steuerung entsperren für Nutzer',
  },
  login: {
    lock: 'Raum sperren (für Nutzer)',
    unlock: 'Raum entsperren (für Nutzer)',
    locked: 'Raum gesperrt (für Nutzer)',
    unlocked: 'Raum entsperrt (für Nutzer)',
    notif_locked: 'Raum gesperrt',
    notif_unlocked: 'Raum entsperrt',
  },
}

export const setting = {
  scroll: 'Scroll-Empfindlichkeit',
  scroll_invert: 'Bildlauf umkehren',
  autoplay: 'Autoplay Video',
  ignore_emotes: 'Emotes ignorieren',
  chat_sound: 'Chat-Sound abspielen',
  keyboard_layout: 'Tastaturbelegung',
  broadcast_title: 'Live-Übertragung',
}

export const connection = {
  reconnecting: 'Erneut verbinden...',
  connected: 'Verbindet',
  disconnected: 'Getrennt',
  kicked: 'Du wurdest aus diesem Raum entfernt.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} hat sich verbunden',
  disconnected: '{name} hat sich getrennt',
  controls_taken: '{name} hat die Steuerung genommen',
  controls_taken_force: 'nahm die Steuerung gewaltsam von',
  controls_taken_steal: 'nahm die Steuerung von {name}',
  controls_released: '{name} hat die Steuerung freigegeben',
  controls_released_force: 'hat die Steuerung gewaltsam losgelassen',
  controls_released_steal: 'hat die Steuerung freigegeben von {name}',
  controls_given: 'hat die Steuerung übergeben an {name}',
  controls_has: '{name} hat die Sterung',
  controls_has_alt: 'Aber ich habe die Person wissen lassen, dass du es wolltest',
  controls_requesting: '{name} fordert die Kontrollen an',
  resolution: 'die Auflösung geändert zu {width}x{height}@{rate}',
  banned: 'sperrte {name}',
  kicked: '{name} wurde rausgeworfen',
}
