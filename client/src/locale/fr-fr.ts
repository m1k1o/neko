export const logout = 'Se déconnecter'
export const unsupported = 'ce navigateur ne prend pas en charge WebRTC'
export const admin_loggedin = "Vous êtes connecté en tant qu'admin"
export const you = 'Vous'
// TODO
//export const somebody = 'Somebody'
export const send_a_message = 'Envoyer un message'

export const side = {
  chat: 'Chat',
  files: 'Fichiers',
  settings: 'Paramètres',
}

export const connect = {
  login_title: 'Veuillez vous connecter',
  invitation_title: 'Vous avez été invité dans cette salle',
  displayname: "Entrez votre nom d'utilisateur",
  password: 'Mot de passe',
  connect: 'Connexion',
  error: 'Erreur de connexion',
  // TODO
  //empty_displayname: 'Display Name cannot be empty.',
}

export const context = {
  ignore: 'Ignorer',
  unignore: 'Ne plus ignorer',
  mute: 'Mute',
  unmute: 'Démute',
  release: 'Forcer le relachement de contrôle',
  take: 'Forcer la prise de contrôle',
  give: 'Donner le contrôle',
  kick: 'Kicker',
  ban: "Bannir l'IP",
  confirm: {
    kick_title: 'Kicker {name}?',
    kick_text: 'Êtes vous sûr de kick {name}?',
    ban_title: 'Bannir {name}?',
    ban_text: 'Voulez-vous bannir {name}? Vous devez relancer le serveur pour annuler le bannissement.',
    mute_title: 'Muter {name}?',
    mute_text: 'Êtes-vous sûr de muter {name}?',
    unmute_title: 'Démute {name}?',
    unmute_text: 'Voulez-vous démuter {name}?',
    button_yes: 'Oui',
    button_cancel: 'Annuler',
  },
}

export const controls = {
  release: 'Relacher le contrôle',
  request: 'Demander le contrôle',
  lock: 'Vérouiller le contrôle',
  unlock: 'Débloquer le contrôle',
  // TODO
  //  has: 'You have control',
  //  hasnot: 'You do not have control',
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
    lock: 'Vérouiller la salle (pour les utilisateurs)',
    unlock: 'Dévérouiller la salle (pour les utilisateurs)',
    locked: 'Salle vérouillée (pour les utilisateurs)',
    unlocked: 'Salle dévérouillée (pour les utilisateurs)',
    notif_locked: 'a vérouillé la salle',
    notif_unlocked: 'a dévérouillé la salle',
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
  scroll: 'Sensibilité de défilement (scroll)',
  scroll_invert: 'Inverser le défilement (scroll)',
  autoplay: 'Jouer automatiquement la vidéo',
  ignore_emotes: 'Ignorer les Emotes',
  chat_sound: 'Jouer le son du tchat',
  keyboard_layout: 'Langue du clavier',
  // TODO
  //broadcast_title: 'Live Broadcast',
}

export const connection = {
  logged_out: 'Vous avez été déconnecté.',
  // TODO
  //reconnecting: 'Reconnecting',
  connected: 'Connecté',
  disconnected: 'Déconnecté',
  // TODO
  //kicked: 'You have been removed from this room.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} connecté',
  disconnected: '{name} déconnecté',
  controls_taken: '{name} a pris le contrôle',
  controls_taken_force: 'a forcé la prise de contrôle',
  controls_taken_steal: 'a pris le contrôle de {name}',
  controls_released: '{name} a relâché le contrôle',
  controls_released_force: 'a forcé la perte de contrôle',
  controls_released_steal: 'a forcé la pêrte de contrôle de {name}',
  controls_given: 'a donné le contrôle à {name}',
  controls_has: '{name} a le contrôle',
  controls_has_alt: "Mais j'ai fait savoir que vous le voulez",
  controls_requesting: '{name} demande le contrôle',
  resolution: 'a changé la résolution pour du {width}x{height}@{rate}',
  banned: 'a banni {name}',
  kicked: 'a kick {name}',
  muted: 'a mute {name}',
  unmuted: 'a démute {name}',
}

export const files = {
  downloads: 'Téléchargements',
  uploads: 'Télécharger',
  upload_here: 'Cliquez ou faites glisser les fichiers ici pour les télécharger',
}
