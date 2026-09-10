export const logout = '登出'
export const unsupported = '你的浏览器不支持 WebRTC'
export const admin_loggedin = '您以管理员身份登录'
export const you = '你'
export const somebody = '某人'
export const send_a_message = '发送消息'

export const ui = {
  live: '直播中',
  about: '关于 n.eko',
  secure: '安全连接',
  github_repository: 'GitHub 仓库',
  remote_browser: '远程浏览器',
  room_eyebrow: 'N.EKO 房间',
  browser_check: '浏览器检查',
  room_controls: '房间控制',
  room_panel: '房间面板',
  toggle_room_panel: '切换房间面板',
  close_room_panel: '关闭房间面板',
  language: '语言',
  enter_fullscreen: '进入全屏',
  change_resolution: '更改分辨率',
  request_or_release_control: '请求或释放控制权',
  open_clipboard: '打开剪贴板',
  picture_in_picture: '画中画',
  open_keyboard: '打开键盘',
  pause: '暂停',
  play: '播放',
  mute: '静音',
  unmute: '取消静音',
  volume: '音量',
}

export const side = {
  chat: '聊天',
  files: '文件',
  settings: '设置',
}

export const chat = {
  empty: '还没有消息，向房间打个招呼吧。',
}

export const connect = {
  login_title: '登录',
  invitation_title: '你已被邀请加入此房间',
  displayname: '您的姓名',
  password: '密码',
  connect: '连接',
  error: '登录错误',
  empty_displayname: '显示名称不能为空',
  show_password: '显示密码',
  hide_password: '隐藏密码',
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
  avatar: '头像',
  avatar_upload: '上传头像',
  avatar_remove: '删除头像',
  avatar_too_large: '头像文件过大',
  avatar_size_limit: '请选择小于 384 KB 的图片。',
  avatar_upload_failed: '头像上传失败',
  avatar_update_failed: '头像更新失败',
  avatar_try_again: '请稍后重试。',
  scroll: '滚动灵敏度',
  scroll_invert: '反转滚动方向',
  autoplay: '自动播放视频',
  ignore_emotes: '忽略表情符号',
  chat_sound: '播放聊天提示音',
  links_in_app: '始终在应用中打开链接',
  keyboard_layout: '键盘布局',
  broadcast_title: '直播流',
  broadcast_placeholder: 'rtmp://a.rtmp.youtube.com/live2/<串流密钥>',
  group_playback: '播放',
  group_chat: '聊天与外观',
  group_input: '输入',
  group_admin: '管理',
  group_session: '会话',
}

export const connection = {
  logged_out: '你已登出',
  connecting: '正在连接',
  reconnecting: '正在重新连接',
  connected: '已连接',
  disconnected: '已断开',
  network_unknown: '网络质量未知',
  network_good: '网络质量良好',
  network_fair: '网络质量一般',
  network_poor: '网络质量较差',
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
  loading: '正在加载文件…',
  empty: '此文件夹为空。',
  downloads: '下载',
  uploads: '上传',
  upload_here: '点击或拖动文件到此处上传',
  delete: '删除',
  delete_title: '删除"{name}"？',
  delete_confirm: '确定要删除此文件吗？',
  select: '选择',
  select_all: '全选',
  unselect_all: '取消全选',
  cancel: '取消',
  delete_selected_title: '删除选中的文件？',
  delete_selected_confirm: '确定要删除选中的 {count} 个文件吗？',
}
