export const logout = 'salir'
export const unsupported = 'este navegador no soporta webrtc'
export const admin_loggedin = 'Registrado como admin'
export const you = 'Tú'
export const send_a_message = 'Enviar un mensaje'

export const side = {
  chat: 'Chat',
  settings: 'Configuración',
}

export const connect = {
  login_title: 'Por favor regístrate',
  invitation_title: 'Te han invitado a esta sala',
  displayname: 'Introduce tu nombre',
  password: 'Contraseña',
  connect: 'Conectar',
  error: 'Error de login',
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
    ban_text: 'Seguroq ue quieres bloquear a {name}? Necesitarás reiniciar el servidor para deshacer esta acción.',
    mute_title: 'Silenciar a {name}?',
    mute_text: 'Seguro que quieres silenciar a {name}?',
    unmute_title: 'Dejar de silenciar a {name}?',
    unmute_text: 'Seguro que quieres dejar de silenciar a {name}?',
    button_yes: 'Sí',
    button_cancel: 'Cancelar',
  },
}

export const controls = {
  release: 'Controles liberador',
  request: 'Controles solicitados',
  lock: 'Controles bloqueados',
  unlock: 'Controles desbloqueados',
}

export const room = {
  lock: 'Bloquear sala (para usuarios)',
  unlock: 'Desbloquear sala (para usuarios)',
  locked: 'Sala bloqueada (para usuarios)',
  unlocked: 'Sala desbloqueada (para usuarios)',
}

export const setting = {
  scroll: 'Sensibilidad del Scroll',
  scroll_invert: 'Invertir Scroll',
  autoplay: 'Auto Reproducir Video',
  ignore_emotes: 'Ignorar Emotes',
  chat_sound: 'Reproducir Sonidos Chat',
  keyboard_layout: 'Keyboard Layout',
  broadcast_is_active: 'Habilitar Broadcast',
  broadcast_url: 'RTMP url',
}

export const connection = {
  logged_out: 'Has salido!',
  // TODO
  //reconnecting: 'Reconnecting',
  connected: 'Connectado correctamente',
  disconnected: 'Has sido desconectado',
  // TODO
  //kicked: 'You have been removed from this room.',
  button_confirm: 'De acuerdo',
}

export const notifications = {
  connected: '{name} se ha conectado',
  disconnected: '{name} se ha desconnectado',
  controls_taken: '{name} tiene los controles',
  controls_taken_force: 'controles confiscados',
  controls_taken_steal: 'cogió los controles de {name}',
  controls_released: '{name} ha liberado los controles',
  controls_released_force: 'controles liberados',
  controls_released_steal: 'controles liberados de {name}',
  controls_given: 'controles asignados a {name}',
  controls_has: '{name} tiene los controles',
  controls_has_alt: 'Pero le diré que quieres los controles',
  controls_requesting: '{name} quiere los controles',
  resolution: 'resolución cambiada a {width}x{height}@{rate}',
  banned: '{name} bloqueado',
  kicked: '{name} expulsado',
  muted: '{name} silenciado',
  unmuted: '{name} no silenciado',
  room_locked: 'bloqueó la sala',
  room_unlocked: 'desbloqueó la sala',
}
