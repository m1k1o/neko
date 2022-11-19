export const logout = '登出'
export const unsupported = '你的浏览器不支持 WebRTC'
export const admin_loggedin = '您以管理钻身份登陆'
export const you = '你'
export const somebody = '某人'
export const send_a_message = '发送消息'

export const side = {
  chat: '聊天',
  files: '文件',
  settings: '设置',
}

export const connect = {
  login_title: '登录',
  invitation_title: '你已被邀请到这个房间',
  displayname: '您的姓名',
  password: '密码',
  connect: '连接',
  error: '登录错误',
  empty_displayname: '显示名称不能为空',
}

export const context = {
  ignore: '忽略',
  unignore: '取消忽略',
  mute: '静音',
  unmute: '取消静音',
  release: '强制释放控制',
  take: '牵制控制',
  give: '给予控制',
  kick: '踢出',
  ban: '禁止 IP',
  confirm: {
    kick_title: '踢出 {name}?',
    kick_text: '你确定你要替 {name}?',
    ban_title: '禁止 {name}?',
    ban_text: '你是否想禁止 {name}? 你将需要重新启动服务器来撤销这一做法.',
    mute_title: '静音 {name}?',
    mute_text: '你确定你要精印吗 {name}?',
    unmute_title: '取消静音 {name}?',
    unmute_text: '你想去下静音吗 {name}?',
    button_yes: '是',
    button_cancel: '取消',
  },
}

export const controls = {
  release: '释放控制',
  request: '请求控制',
  lock: '锁定控制',
  unlock: '解锁控制',
  has: '你有控制',
  hasnot: '你没有控制',
}

export const locks = {
  control: {
    lock: '对所有用户进行锁定控制',
    unlock: '对所有用户进行解锁控制',
    locked: '锁定的控制装置',
    unlocked: '解锁的控制装置',
    notif_locked: '为用户锁定控制',
    notif_unlocked: '为用户解锁控制',
  },
  login: {
    lock: '所有用户的锁定室',
    unlock: '所有用户的解锁室',
    locked: '为所有用户锁定的房间',
    unlocked: '为所有用户解锁的房间',
    notif_locked: '锁上房间',
    notif_unlocked: '解锁房间',
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
  scroll: '滚动敏感度',
  scroll_invert: '反转滚动敏感度',
  autoplay: '自动播放视频',
  ignore_emotes: '忽略表情符号',
  chat_sound: '播放聊天声音',
  keyboard_layout: '键盘布局',
  broadcast_title: '现场流媒体',
}

export const connection = {
  logged_out: '你已登出',
  reconnecting: '正在重新连接',
  connected: '已连接',
  disconnected: '已断开',
  kicked: '你已被踢出',
  button_confirm: '好的',
}

export const notifications = {
  connected: '{name} 已连接',
  disconnected: '{name} 已断开',
  controls_taken: '{name} 采取了控制',
  controls_taken_force: '被迫接受控制',
  controls_taken_steal: '掌握控制权 从 {name}',
  controls_released: '{name} 释放控制',
  controls_released_force: '强制解除控制',
  controls_released_steal: '强制解除控制 从 {name}',
  controls_given: '将控制权交给 {name}',
  controls_has: '{name} 拥有控制权',
  controls_has_alt: '但我让那个人知道你想要它',
  controls_requesting: '{name} 正在请求控制',
  resolution: '将分辨率改为 {width}x{height}@{rate}',
  banned: '被禁止 {name}',
  kicked: '被踢的 {name}',
  muted: '鸟粪 {name}',
  unmuted: '取消静音 {name}',
}

export const files = {
  downloads: '下载',
  uploads: '上传',
  upload_here: '点击或拖动文件到这里来上传',
}
