export const logout = 'ログアウト'
export const unsupported = 'このウェブブラウザは WebRTC をサポートしていません'
export const admin_loggedin = '管理者としてログインしています'
export const you = 'あなた'
export const somebody = '誰か'
export const send_a_message = 'メッセージを送信'

export const side = {
  chat: 'チャット',
  files: 'ファイル',
  settings: '設定',
}

export const connect = {
  login_title: 'ログインしてください',
  invitation_title: 'このルームに招待されました',
  displayname: 'あなたの表示名を入力してください',
  password: 'パスワード',
  connect: '接続',
  error: 'ログインエラー',
  empty_displayname: '表示名は空欄にはできません',
}

export const context = {
  ignore: '無視する',
  unignore: '無視を解除',
  mute: 'ミュート',
  unmute: 'ミュート解除',
  release: '強制的にコントロールを解放する',
  take: '強制的にコントロールを得る',
  give: 'コントロールを譲渡する',
  kick: '追い出す',
  ban: 'IPを禁止にする',
  confirm: {
    kick_title: '{name} を追い出しますか?',
    kick_text: '本当に {name} を追い出しますか?',
    ban_title: '{name} を禁止にしますか?',
    ban_text: '本当に {name} を禁止にしますか? 取り消すにはサーバを再起動する必要があります',
    mute_title: '{name} をミュートしますか?',
    mute_text: '本当に {name} をミュートしますか?',
    unmute_title: '{name} のミュートを解除しますか?',
    unmute_text: '{name} のミュートを解除しますか?',
    button_yes: 'はい',
    button_cancel: 'キャンセル',
  },
}

export const controls = {
  release: 'コントロールを解放',
  request: 'コントロールを要求',
  lock: 'コントロールをロック',
  unlock: 'コントロールのロックを解除',
  has: 'あなたにコントロールがあります',
  hasnot: 'あなたにはコントロールがありません',
}

export const locks = {
  control: {
    lock: 'コントロールをロック (ユーザに対して)',
    unlock: 'コントロールのロックを解除 (ユーザに対して)',
    locked: 'コントロールはロックされています (ユーザに対して)',
    unlocked: 'コントロールはロックされていません (ユーザに対して)',
    notif_locked: 'ユーザに対してコントロールをロックしました',
    notif_unlocked: 'ユーザに対してコントロールのロックを解除しました',
  },
  login: {
    lock: 'ルームをロック (ユーザに対して)',
    unlock: 'ルームのロックを解除 (ユーザに対して)',
    locked: 'ルームはロックされています (ユーザに対して)',
    unlocked: 'ルームはロックされていません (ユーザに対して)',
    notif_locked: 'ルームをロックしました',
    notif_unlocked: 'ルームのロックを解除しました',
  },
  file_transfer: {
    lock: 'ファイル転送をロック (ユーザに対して)',
    unlock: 'ファイル転送のロックを解除 (ユーザに対して)',
    locked: 'ファイル転送はロックされています (ユーザに対して)',
    unlocked: 'ファイル転送はロックされていません (ユーザに対して)',
    notif_locked: 'ファイル転送をロックしました',
    notif_unlocked: 'ファイル転送のロックを解除しました',
  },
}

export const setting = {
  scroll: 'スクロールの感度',
  scroll_invert: 'スクロールを反転する',
  autoplay: '動画を自動再生する',
  ignore_emotes: '絵文字を無視する',
  chat_sound: 'チャットで音を再生する',
  keyboard_layout: 'キーボード配列',
  broadcast_title: 'ライブ配信',
}

export const connection = {
  logged_out: 'ログアウトしました',
  reconnecting: '再接続中...',
  connected: '接続しました',
  disconnected: '切断しました',
  kicked: 'あなたはこの部屋から追い出されました',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} が接続しました',
  disconnected: '{name} が切断しました',
  controls_taken: '{name} がコントロールを得ました',
  controls_taken_force: 'がコントロールを強制的に得ました',
  controls_taken_steal: 'が {name} からコントロールを得ました',
  controls_released: '{name} がコントロールを解放しました',
  controls_released_force: 'が強制的にコントロールを解放しました',
  controls_released_steal: 'が {name} からコントロールを解放しました',
  controls_given: '{name} にコントロールを譲渡しました',
  controls_has: '{name} にコントロールがあります',
  controls_has_alt: 'しかし、その人にあなたがそれを希望していることを伝えました',
  controls_requesting: '{name} がコントロールを要求しています',
  resolution: '解像度を {width}x{height}@{rate} に変更しました',
  banned: '{name} を禁止しました',
  kicked: '{name} を追い出しました',
  muted: '{name} をミュートにしました',
  unmuted: '{name} のミュートを解除しました',
}

export const files = {
  downloads: 'ダウンロード',
  uploads: 'アップロード',
  upload_here: 'アップロードするにはここをクリックするかファイルをドラッグしてください',
}
