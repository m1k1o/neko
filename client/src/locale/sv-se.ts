export const logout = 'logga ut'
export const unsupported = 'denna webbläsare har inte stöd för webrtc'
export const admin_loggedin = 'Du är inloggad som en administratör'
export const you = 'Du'
export const somebody = 'Någon' // TODO: kontrollera översättning
export const send_a_message = 'Skicka ett meddelande'

export const side = {
  chat: 'Chatt',
  files: 'Filer',
  settings: 'Inställningar',
}

export const connect = {
  login_title: 'Vänligen logga in',
  invitation_title: 'Du har blivit inbjuden till detta rum',
  displayname: 'Skriv in ditt namn',
  password: 'Lösenord',
  connect: 'Anslut',
  error: 'Inloggningsfel',
  empty_displayname: 'Visningsnamn kan inte vara tomt.', // TODO: kontrollera översättning
}

export const context = {
  ignore: 'Ignorera',
  unignore: 'Inte ignorera',
  mute: 'Tysta',
  unmute: 'Ta bort tystning',
  release: 'Tvinga ta bort kontrollen',
  take: 'Tvinga ta kontrollen',
  give: 'Ge kontrollen',
  kick: 'Sparka',
  ban: 'Bannlys IP',
  confirm: {
    kick_title: 'Sparka {name}?',
    kick_text: 'Är du säker du vill sparka {name}?',
    ban_title: 'Bannlys {name}?',
    ban_text: 'Är du säker du vill bannlysa {name}? Du behöver starta om servern för att ta bort den bannlysningen.',
    mute_title: 'Tysta {name}?',
    mute_text: 'Är du säker du vill tysta {name}?',
    unmute_title: 'Ta bort tystningen {name}?',
    unmute_text: 'Är du säker du vill ta bort tystningen {name}?',
    button_yes: 'Ja',
    button_cancel: 'Avbryt',
  },
}

export const controls = {
  release: 'Ta kontrollen',
  request: 'Fråga om kontroll',
  lock: 'Lås kontrollen',
  unlock: 'Lås upp kontrollen',
  has: 'Du har kontrollen', // TODO: kontrollera översättning
  hasnot: 'Du har inte kontrollen', // TODO: kontrollera översättning
}

export const locks = {
  control: {
    lock: 'Lås kontroller (för användare)',
    unlock: 'Lås upp kontroller (för användare)',
    locked: 'Kontroller låsta (för användare)',
    unlocked: 'Kontroller upplåsta (för användare)',
    notif_locked: 'låste kontroller för användare',
    notif_unlocked: 'låste upp kontroller för användare',
  },
  login: {
    lock: 'Lås rum (för användare)',
    unlock: 'Lås upp rummet (för användare)',
    locked: 'Rum låst (för användare)',
    unlocked: 'Rum upplåst (för användare)',
    notif_locked: 'låste rummet',
    notif_unlocked: 'låste upp rummet',
  },
  file_transfer: {
    lock: 'Lås filöverföring (för användare)',
    unlock: 'Lås upp filöverföring (för användare)',
    locked: 'Filöverföring låst (för användare)',
    unlocked: 'Filöverföring upplåst (för användare)',
    notif_locked: 'låste filöverföring',
    notif_unlocked: 'låste upp filöverföring',
  },
}

export const setting = {
  scroll: 'Scrollkänslighet',
  scroll_invert: 'Vänd scrollen',
  autoplay: 'Automatisk uppspelning av video',
  ignore_emotes: 'Ignorera emojis',
  chat_sound: 'Spela chattljud',
  keyboard_layout: 'Tangentbordslayout',
  broadcast_title: 'Livesändning',
}

export const connection = {
  logged_out: 'Du har blivit utloggad.',
  reconnecting: 'Återansluter...',
  connected: 'Ansluten',
  disconnected: 'Frånkopplad',
  kicked: 'Du har blivit borttagen från detta rum.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} anslöt',
  disconnected: '{name} kopplade ifrån',
  controls_taken: '{name} tog kontrollen',
  controls_taken_force: 'tvinga ta kontrollen',
  controls_taken_steal: 'tog kontrollen från {name}',
  controls_released: '{name} lämnade kontrollen',
  controls_released_force: 'tvingade ta bort kontrollen',
  controls_released_steal: 'tog bort kontrollen från {name}',
  controls_given: 'gav kontrollen till {name}',
  controls_has: '{name} har kontrollen',
  controls_has_alt: 'Men jag låter dem veta att du vill ha den',
  controls_requesting: '{name} frågar om kontrollen',
  resolution: 'ändrade upplösningen till {width}x{height}@{rate}',
  banned: 'bannlyste {name}',
  kicked: 'sparkade {name}',
  muted: 'tystade {name}',
  unmuted: 'tog bort tystningen på {name}',
}

export const files = {
  downloads: 'Nedladdningar',
  uploads: 'Ladda upp',
  upload_here: 'Klicka eller dra filer hit för att ladda upp dem',
}
