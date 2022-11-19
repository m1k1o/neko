export const logout = 'logga ut'
export const unsupported = 'denna webbläsare har inte stöd för webrtc'
export const admin_loggedin = 'Du är inloggad som en administratör'
export const you = 'Du'
// TODO
//export const somebody = 'Somebody'
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
  // TODO
  //empty_displayname: 'Display Name cannot be empty.',
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
    lock: 'Lås rum (för användare)',
    unlock: 'Lås upp rummet (för användare)',
    locked: 'Rum låst (för användare)',
    unlocked: 'Rum upplåst (för användare)',
    notif_locked: 'låste rummet',
    notif_unlocked: 'låste upp rummet',
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
  scroll: 'Scrollkänslighet',
  scroll_invert: 'Vänd Scrollen',
  autoplay: 'Automatisk uppspelning av Video',
  ignore_emotes: 'Ignorera Emotes',
  chat_sound: 'Spela Chatt Ljud',
  keyboard_layout: 'Tangentbordslayout',
  // TODO
  //broadcast_title: 'Live Broadcast',
}

export const connection = {
  logged_out: 'Du har blivit utloggad!',
  // TODO
  //reconnecting: 'Reconnecting',
  connected: 'Du har loggats in',
  disconnected: 'Du har blivit frånkopplad',
  // TODO
  //kicked: 'You have been removed from this room.',
  button_confirm: 'Ok',
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
