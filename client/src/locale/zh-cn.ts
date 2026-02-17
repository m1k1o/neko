export const logout = '登出'
export const unsupported = '你的浏览器不支持 WebRTC'
export const admin_loggedin = '您以管理员身份登录'
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
  invitation_title: '你已被邀请加入此房间',
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
  take: '强制接管控制',
  give: '移交控制',
  kick: '踢出',
  ban: '封禁 IP',
  confirm: {
    kick_title: '踢出 {name}?',
    kick_text: '你确定要踢出 {name} 吗?',
    ban_title: '封禁 {name}?',
    ban_text: '你确定要封禁 {name} 吗？如需解除封禁需重启服务。',
    mute_title: '静音 {name}?',
    mute_text: '你确定要将 {name} 静音吗?',
    unmute_title: '取消静音 {name}?',
    unmute_text: '你确定要取消 {name} 的静音吗?',
    button_yes: '是',
    button_cancel: '取消',
  },
}

export const controls = {
  release: '释放控制',
  request: '请求控制',
  lock: '锁定控制',
  unlock: '解锁控制',
  has: '你拥有控制权',
  hasnot: '你没有控制权',
}

export const locks = {
  control: {
    lock: '锁定所有用户的控制',
    unlock: '解锁所有用户的控制',
    locked: '控制已锁定',
    unlocked: '控制已解锁',
    notif_locked: '已为用户锁定控制',
    notif_unlocked: '已为用户解锁控制',
  },
  login: {
    lock: '锁定所有用户的房间',
    unlock: '解锁所有用户的房间',
    locked: '房间已为所有用户锁定',
    unlocked: '房间已为所有用户解锁',
    notif_locked: '房间已锁定',
    notif_unlocked: '房间已解锁',
  },
  file_transfer: {
   lock: '锁定文件传输（对用户）',
   unlock: '解锁文件传输（对用户）',
   locked: '文件传输已锁定（对用户）',
   unlocked: '文件传输已解锁（对用户）',
   notif_locked: '已锁定文件传输',
   notif_unlocked: '已解锁文件传输',
  },
}

export const setting = {
  scroll: '滚动灵敏度',
  scroll_invert: '反转滚动方向',
  autoplay: '自动播放视频',
  ignore_emotes: '忽略表情符号',
  chat_sound: '播放聊天提示音',
  keyboard_layout: '键盘布局',
  broadcast_title: '直播流',
}

export const connection = {
  logged_out: '你已登出',
  reconnecting: '正在重新连接',
  connected: '已连接',
  disconnected: '已断开',
  kicked: '你已被踢出',
  button_confirm: '确定',
}

export const notifications = {
  connected: '{name} 已连接',
  disconnected: '{name} 已断开',
  controls_taken: '{name} 获得了控制权',
  controls_taken_force: '强制获得控制权',
  controls_taken_steal: '从 {name} 夺取了控制权',
  controls_released: '{name} 释放了控制权',
  controls_released_force: '强制释放控制权',
  controls_released_steal: '从 {name} 强制释放控制权',
  controls_given: '将控制权交给了 {name}',
  controls_has: '{name} 拥有控制权',
  controls_has_alt: '但我已通知对方你想要控制权',
  controls_requesting: '{name} 正在请求控制权',
  resolution: '分辨率已更改为 {width}x{height}@{rate}',
  banned: '{name} 已被封禁',
  kicked: '{name} 已被踢出',
  muted: '{name} 已被静音',
  unmuted: '{name} 已取消静音',
}

export const files = {
  downloads: '下载',
  uploads: '上传',
  upload_here: '点击或拖动文件到此处上传',
}
