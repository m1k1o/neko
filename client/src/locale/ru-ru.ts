export const logout = 'выход'
export const unsupported = 'этот браузер не поддерживает WebRTC'
export const admin_loggedin = 'Вы вошли как админ'
export const you = 'Вы'
export const somebody = 'Кто-то'
export const send_a_message = 'Отправить сообщение'

export const side = {
  chat: 'Чат',
  files: 'Файлы',
  settings: 'Настройки',
}

export const connect = {
  login_title: 'Пожалуйста, войдите',
  invitation_title: 'Вас пригласили в эту комнату',
  displayname: 'Введите ваше отображаемое имя',
  password: 'Пароль',
  connect: 'Подключиться',
  error: 'Ошибка входа',
  empty_displayname: 'Отображаемое имя не может быть пустым.',
}

export const context = {
  ignore: 'Игнорировать',
  unignore: 'Не игнорировать',
  mute: 'Заглушить',
  unmute: 'Перестать глушить',
  release: 'Принудительно освободить управление',
  take: 'Принудительно взять управление',
  give: 'Дать управление',
  kick: 'Выкинуть',
  ban: 'Забанить IP',
  confirm: {
    kick_title: 'Выкинуть {name}?',
    kick_text: 'Вы уверены, что хотите выкинуть {name}?',
    ban_title: 'Забанить {name}?',
    ban_text: 'Вы хотите забанить {name}? Для отмены придётся перезапустить сервер.',
    mute_title: 'Заглушить {name}?',
    mute_text: 'Вы уверены, что хотите заглушить {name}?',
    unmute_title: 'Перестать глушить {name}?',
    unmute_text: 'Вы хотите перестать глушить {name}?',
    button_yes: 'Да',
    button_cancel: 'Отмена',
  },
}

export const controls = {
  release: 'Освободить управление',
  request: 'Запросить управление',
  lock: 'Закрепить управление',
  unlock: 'Открепить управление',
  has: 'Управление у вас',
  hasnot: 'Вы не управляете',
}

export const locks = {
  control: {
    lock: 'Закрепить управление (для пользователей)',
    unlock: 'Открепить управление (для пользователей)',
    locked: 'Управление закреплено (для пользователей)',
    unlocked: 'Управление откреплено (для пользователей)',
    notif_locked: 'закреплено управление для пользователей',
    notif_unlocked: 'откреплено управление для пользователей',
  },
  login: {
    lock: 'Закрыть комнату (для пользователей)',
    unlock: 'Открыть комнату (для пользователей)',
    locked: 'Комната закрыта (для пользователей)',
    unlocked: 'Комната открыта (для пользователей)',
    notif_locked: 'комната закрыта',
    notif_unlocked: 'комната открыта',
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
  scroll: 'Чувствительность прокрутки',
  scroll_invert: 'Инвертировать прокрутку',
  autoplay: 'Автовоспроизведение видео',
  ignore_emotes: 'Игнорировать эмоции',
  chat_sound: 'Проигрывать звук чата',
  keyboard_layout: 'Раскладка клавиатуры',
  broadcast_title: 'Прямой эфир',
}

export const connection = {
  logged_out: 'Вы вышли.',
  reconnecting: 'Переподключение...',
  connected: 'Подключено',
  disconnected: 'Отключено',
  kicked: 'Вас выкинули из комнаты.',
  button_confirm: 'ОК',
}

export const notifications = {
  connected: '{name} подключился',
  disconnected: '{name} отключился',
  controls_taken: '{name} взял управление',
  controls_taken_force: 'взял управление принудительно',
  controls_taken_steal: 'взял управление у {name}',
  controls_released: '{name} освободил управление',
  controls_released_force: 'освободил управление принудительно',
  controls_released_steal: 'освободил управление у {name}',
  controls_given: 'дал управление {name}',
  controls_has: '{name} теперь управляет',
  controls_has_alt: 'Мы уведомим пользователя о запросе',
  controls_requesting: '{name} запрашивает управление',
  resolution: 'разрешение изменено на {width}x{height}@{rate}',
  banned: 'забанен {name}',
  kicked: 'выкинут {name}',
  muted: 'заглушен {name}',
  unmuted: 'не заглушен {name}',
}

export const files = {
  downloads: 'Загрузки',
  uploads: 'Загрузить',
  upload_here: 'Нажмите или перетащите сюда файлы для загрузки',
}
