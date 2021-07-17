export const logout = 'log out'
export const unsupported = 'this web-browser does not support WebRTC'
export const admin_loggedin = 'You are logged in as an admin'
export const you = 'You'
export const send_a_message = 'Send a message'

export const side = {
  chat: 'Chat',
  settings: 'Settings',
}

export const connect = {
  login_title: 'Please Log In',
  invitation_title: 'You have been invited to this room',
  displayname: 'Enter your display name',
  password: 'Password',
  connect: 'Connect',
  error: 'Login error',
}

export const context = {
  ignore: 'Ignore',
  unignore: 'Unignore',
  mute: 'Mute',
  unmute: 'Unmute',
  release: 'Force Release Controls',
  take: 'Force Take Controls',
  give: 'Give Controls',
  kick: 'Kick',
  ban: 'Ban IP',
  confirm: {
    kick_title: 'Kick {name}?',
    kick_text: 'Are you sure you want to kick {name}?',
    ban_title: 'Ban {name}?',
    ban_text: 'Do you want to ban {name}? You will need to restart the server to undo this.',
    mute_title: 'Mute {name}?',
    mute_text: 'Are you sure you want to mute {name}?',
    unmute_title: 'Unmute {name}?',
    unmute_text: 'Do you want to unmute {name}?',
    button_yes: 'Yes',
    button_cancel: 'Cancel',
  },
}

export const controls = {
  release: 'Release Controls',
  request: 'Request Controls',
  lock: 'Lock Controls',
  unlock: 'Unlock Controls',
}

export const room = {
  lock: 'Lock Room (for users)',
  unlock: 'Unlock Room (for users)',
  locked: 'Room Locked (for users)',
  unlocked: 'Room Unlocked (for users)',
}

export const setting = {
  scroll: 'Scroll Sensitivity',
  scroll_invert: 'Invert Scroll',
  autoplay: 'Autoplay Video',
  ignore_emotes: 'Ignore Emotes',
  chat_sound: 'Play Chat Sound',
  keyboard_layout: 'Keyboard Layout',
  broadcast_is_active: 'Broadcast Enabled',
  broadcast_url: 'RTMP URL',
}

export const connection = {
  logged_out: 'You have been logged out.',
  reconnecting: 'Reconnecting...',
  connected: 'Connected',
  disconnected: 'Disconnected',
  kicked: 'You have been removed from this room.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} connected',
  disconnected: '{name} disconnected',
  controls_taken: '{name} took the controls',
  controls_taken_force: 'took the controls forcibly',
  controls_taken_steal: 'took the controls from {name}',
  controls_released: '{name} released the controls',
  controls_released_force: 'released the controls forcibly',
  controls_released_steal: 'released the controls from {name}',
  controls_given: 'gave the controls to {name}',
  controls_has: '{name} has the controls',
  controls_has_alt: 'But I let the person know you wanted it',
  controls_requesting: '{name} is requesting the controls',
  resolution: 'changed the resolution to {width}x{height}@{rate}',
  banned: 'banned {name}',
  kicked: 'kicked {name}',
  muted: 'muted {name}',
  unmuted: 'unmuted {name}',
  room_locked: 'locked the room',
  room_unlocked: 'unlocked the room',
}
