export const logout = 'Wyloguj się'
export const unsupported = 'Ta przeglądarka nie obsługuje WebRTC'
export const admin_loggedin = 'Jesteś zalogowany jako administrator'
export const you = 'Ty'
export const somebody = 'Ktoś'
export const send_a_message = 'Wyślij wiadomość'

export const side = {
  chat: 'Czat',
  files: 'Pliki',
  settings: 'Ustawienia',
}

export const connect = {
  login_title: 'Zaloguj się',
  invitation_title: 'Otrzymałeś zaproszenie do tego pokoju',
  displayname: 'Podaj swoją nazwę',
  password: 'Hasło',
  connect: 'Połącz',
  error: 'Błąd logowania',
  empty_displayname: 'Pole nazwy użytkownika nie może być puste.',
}

export const context = {
  ignore: 'Ignoruj',
  unignore: 'Przestań ignorować',
  mute: 'Wycisz',
  unmute: 'Cofnij wyciszenie',
  release: 'Wymuś zwolnienie sterowania',
  take: 'Wymuś przejęcie sterowania',
  give: 'Przekaż sterowanie',
  kick: 'Wyrzuć',
  ban: 'Zbanuj IP',
  confirm: {
    kick_title: 'Wyrzucić {name}?',
    kick_text: 'Czy na pewno chcesz wyrzucić {name}?',
    ban_title: 'Zbanować {name}?',
    ban_text: 'Czy chcesz zbanować {name}? Aby cofnąć, musisz zrestartować serwer.',
    mute_title: 'Wyciszyć {name}?',
    mute_text: 'Czy na pewno chcesz wyciszyć {name}?',
    unmute_title: 'Cofnąć wyciszenie {name}?',
    unmute_text: 'Czy chcesz cofnąć wyciszenie {name}?',
    button_yes: 'Tak',
    button_cancel: 'Anuluj',
  },
}

export const controls = {
  release: 'Zwolnij sterowanie',
  request: 'Poproś o sterowanie',
  lock: 'Zablokuj sterowanie',
  unlock: 'Odblokuj sterowanie',
  has: 'Masz sterowanie',
  hasnot: 'Nie masz sterowania',
}

export const locks = {
  control: {
    lock: 'Zablokuj sterowanie (dla użytkowników)',
    unlock: 'Odblokuj sterowanie (dla użytkowników)',
    locked: 'Sterowanie zablokowane (dla użytkowników)',
    unlocked: 'Sterowanie odblokowane (dla użytkowników)',
    notif_locked: 'zablokował sterowanie dla użytkowników',
    notif_unlocked: 'odblokował sterowanie dla użytkowników',
  },
  login: {
    lock: 'Zablokuj pokój (dla użytkowników)',
    unlock: 'Odblokuj pokój (dla użytkowników)',
    locked: 'Pokój zablokowany (dla użytkowników)',
    unlocked: 'Pokój odblokowany (dla użytkowników)',
    notif_locked: 'zablokował pokój',
    notif_unlocked: 'odblokował pokój',
  },
  file_transfer: {
    lock: 'Zablokuj transfer plików (dla użytkowników)',
    unlock: 'Odblokuj transfer plików (dla użytkowników)',
    locked: 'Transfer plików zablokowany (dla użytkowników)',
    unlocked: 'Transfer plików odblokowany (dla użytkowników)',
    notif_locked: 'zablokował transfer plików',
    notif_unlocked: 'odblokował transfer plików',
  },
}

export const setting = {
  scroll: 'Czułość przewijania',
  scroll_invert: 'Odwróć przewijanie',
  autoplay: 'Autoodtwarzanie wideo',
  ignore_emotes: 'Ignoruj emotki',
  chat_sound: 'Odtwarzaj dźwięk czatu',
  keyboard_layout: 'Układ klawiatury',
  broadcast_title: 'Transmisja na żywo',
}

export const connection = {
  logged_out: 'Zostałeś wylogowany.',
  reconnecting: 'Ponowne łączenie...',
  connected: 'Połączono',
  disconnected: 'Rozłączono',
  kicked: 'Zostałeś usunięty z pokoju.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} dołączył',
  disconnected: '{name} opuścił',
  controls_taken: '{name} przejął sterowanie',
  controls_taken_force: 'przejął sterowanie siłą',
  controls_taken_steal: 'przejął sterowanie od {name}',
  controls_released: '{name} zwolnił sterowanie',
  controls_released_force: 'zwolnił sterowanie siłą',
  controls_released_steal: 'zwolnił sterowanie od {name}',
  controls_given: 'przekazał sterowanie {name}',
  controls_has: '{name} ma sterowanie',
  controls_has_alt: 'Ale poinformowałem osobę, że chciałeś je otrzymać',
  controls_requesting: '{name} prosi o sterowanie',
  resolution: 'zmienił rozdzielczość na {width}x{height}@{rate}',
  banned: 'zbanował {name}',
  kicked: 'wyrzucił {name}',
  muted: 'wyciszył {name}',
  unmuted: 'cofnął wyciszenie {name}',
}

export const files = {
  downloads: 'Pobrane pliki',
  uploads: 'Wysyłane pliki',
  upload_here: 'Kliknij lub przeciągnij pliki tutaj, aby przesłać',
}
