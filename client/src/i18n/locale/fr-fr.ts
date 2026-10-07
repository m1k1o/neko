export const logout = 'Se déconnecter'
export const unsupported = 'ce navigateur ne prend pas en charge WebRTC'
export const admin_loggedin = "Vous êtes connecté en tant qu'admin"
export const you = 'Vous'
export const somebody = "Quelqu'un" // TODO: vérifier la traduction

export const side = {
  settings: 'Paramètres',
}

export const connect = {
  login_title: 'Veuillez vous connecter',
  invitation_title: 'Vous avez été invité dans cette salle',
  displayname: "Entrez votre nom d'utilisateur",
  password: 'Mot de passe',
  connect: 'Connexion',
  error: 'Erreur de connexion',
  empty_displayname: "Le nom d'affichage ne peut pas être vide.", // TODO: vérifier la traduction
}

export const context = {
  ignore: 'Ignorer',
  unignore: 'Ne plus ignorer',
  release: 'Forcer le relachement de contrôle',
  take: 'Forcer la prise de contrôle',
  give: 'Donner le contrôle',
  kick: 'Kicker',
  confirm: {
    kick_title: 'Kicker {name}?',
    kick_text: 'Êtes vous sûr de kick {name}?',
    ban_title: 'Bannir {name}?',
    button_yes: 'Oui',
    button_cancel: 'Annuler',
  },
}

export const controls = {
  release: 'Relacher le contrôle',
  request: 'Demander le contrôle',
  lock: 'Vérouiller le contrôle',
  unlock: 'Débloquer le contrôle',
  has: 'Vous avez le contrôle', // TODO: vérifier la traduction
  hasnot: "Vous n'avez pas le contrôle", // TODO: vérifier la traduction
}

export const locks = {
  control: {
    lock: 'Verrouiller les contrôles (pour les utilisateurs)',
    unlock: 'Déverrouiller les contrôles (pour les utilisateurs)',
    locked: 'Contrôles verrouillés (pour les utilisateurs)',
    unlocked: 'Contrôles déverrouillés (pour les utilisateurs)',
    notif_locked: 'contrôles verrouillés pour les utilisateurs',
    notif_unlocked: 'contrôles déverrouillés pour les utilisateurs',
  },
  login: {
    lock: 'Vérouiller la salle (pour les utilisateurs)',
    unlock: 'Dévérouiller la salle (pour les utilisateurs)',
    locked: 'Salle vérouillée (pour les utilisateurs)',
    unlocked: 'Salle dévérouillée (pour les utilisateurs)',
    notif_locked: 'a vérouillé la salle',
    notif_unlocked: 'a dévérouillé la salle',
  },
}

export const setting = {
  scroll: 'Sensibilité de défilement',
  scroll_invert: 'Inverser le défilement',
  autoplay: 'Lecture automatique de la vidéo',
  ignore_emotes: 'Ignorer les émoticônes',
  chat_sound: 'Son du chat',
  keyboard_layout: 'Disposition du clavier',
  broadcast_title: 'Diffusion en direct',
}

export const connection = {
  reconnecting: 'Reconnexion...',
  connected: 'Connecté',
  disconnected: 'Déconnecté',
  kicked: 'Vous avez été expulsé de cette salle.',
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
}
