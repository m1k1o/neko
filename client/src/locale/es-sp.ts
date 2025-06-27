export const logout = 'salir'
export const unsupported = 'este navegador no soporta webrtc'
export const admin_loggedin = 'Registrado como admin'
export const you = 'Tú'
export const somebody = 'Alguien'
export const send_a_message = 'Enviar un mensaje'

export const side = {
  chat: 'Chat',
  files: 'Archivos',
  settings: 'Configuración',
}

export const connect = {
  login_title: 'Por favor regístrate',
  invitation_title: 'Te han invitado a esta sala',
  displayname: 'Introduce tu nombre',
  password: 'Contraseña',
  connect: 'Conectar',
  error: 'Error de login',
  empty_displayname: 'El nombre para mostrar no puede estar vacío.',
}

export const context = {
  ignore: 'Ignorar',
  unignore: 'No ignorar',
  mute: 'Silenciar',
  unmute: 'No silenciar',
  release: 'Forzar liberar los controles',
  take: 'Forzar obtener los controles',
  give: 'Dar los controles',
  kick: 'Echar',
  ban: 'Bloquear IP',
  confirm: {
    kick_title: 'Echar a {name}?',
    kick_text: 'Seguro que quiere echar a {name}?',
    ban_title: 'Bloquear a {name}?',
    ban_text: '¿Seguro que quieres bloquear a {name}? Necesitarás reiniciar el servidor para deshacer esta acción.',
    mute_title: 'Silenciar a {name}?',
    mute_text: 'Seguro que quieres silenciar a {name}?',
    unmute_title: 'Dejar de silenciar a {name}?',
    unmute_text: 'Seguro que quieres dejar de silenciar a {name}?',
    button_yes: 'Sí',
    button_cancel: 'Cancelar',
  },
}

export const controls = {
  release: 'Liberar controles',
  request: 'Solicitar controles',
  lock: 'Bloquear controles',
  unlock: 'Desbloquear controles',
  has: 'Tienes el control',
  hasnot: 'No tienes el control',
}

export const locks = {
  control: {
    lock: 'Bloquear controles (para usuarios)',
    unlock: 'Desbloquear controles (para usuarios)',
    locked: 'Controles bloqueados (para usuarios)',
    unlocked: 'Controles desbloqueados (para usuarios)',
    notif_locked: 'controles bloqueados para usuarios',
    notif_unlocked: 'controles desbloqueados para usuarios',
  },
  login: {
    lock: 'Bloquear sala (para usuarios)',
    unlock: 'Desbloquear sala (para usuarios)',
    locked: 'Sala bloqueada (para usuarios)',
    unlocked: 'Sala desbloqueada (para usuarios)',
    notif_locked: 'bloqueó la sala',
    notif_unlocked: 'desbloqueó la sala',
  },
  file_transfer: {
    lock: 'Bloquear transferencia de archivos (para usuarios)',
    unlock: 'Desbloquear transferencia de archivos (para usuarios)',
    locked: 'Transferencia de archivos bloqueada (para usuarios)',
    unlocked: 'Transferencia de archivos desbloqueada (para usuarios)',
    notif_locked: 'transferencia de archivos bloqueada',
    notif_unlocked: 'transferencia de archivos desbloqueada',
  },
}

export const setting = {
  scroll: 'Sensibilidad del scroll',
  scroll_invert: 'Invertir scroll',
  autoplay: 'Reproducir video automáticamente',
  ignore_emotes: 'Ignorar emoticonos',
  chat_sound: 'Reproducir sonido del chat',
  keyboard_layout: 'Diseño del teclado',
  broadcast_title: 'Transmisión en vivo',
}

export const connection = {
  logged_out: 'Has sido desconectado.',
  reconnecting: 'Reconectando...',
  connected: 'Conectado',
  disconnected: 'Desconectado',
  kicked: 'Has sido expulsado de esta sala.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} se conectó',
  disconnected: '{name} se desconectó',
  controls_taken: '{name} tomó los controles',
  controls_taken_force: 'tomó los controles por la fuerza',
  controls_taken_steal: 'tomó los controles de {name}',
  controls_released: '{name} liberó los controles',
  controls_released_force: 'liberó los controles por la fuerza',
  controls_released_steal: 'liberó los controles de {name}',
  controls_given: 'dio los controles a {name}',
  controls_has: '{name} tiene los controles',
  controls_has_alt: 'Pero le hice saber a la persona que los querías',
  controls_requesting: '{name} está solicitando los controles',
  resolution: 'cambió la resolución a {width}x{height}@{rate}',
  banned: 'bloqueó a {name}',
  kicked: 'expulsó a {name}',
  muted: 'silenció a {name}',
  unmuted: 'quitó el silencio a {name}',
}

export const files = {
  downloads: 'Descargas',
  uploads: 'Subidas',
  upload_here: 'Haz clic o arrastra archivos aquí para subirlos',
}
