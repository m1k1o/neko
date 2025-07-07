export const logout = 'keluar'
export const unsupported = 'peramban web ini tidak mendukung WebRTC'
export const admin_loggedin = 'Anda masuk sebagai admin'
export const you = 'Anda'
export const somebody = 'Seseorang'
export const send_a_message = 'Kirim pesan'

export const side = {
  chat: 'Obrolan',
  files: 'Berkas',
  settings: 'Pengaturan',
}

export const connect = {
  login_title: 'Silakan Masuk',
  invitation_title: 'Anda telah diundang ke ruangan ini',
  displayname: 'Masukkan nama tampilan Anda',
  password: 'Kata Sandi',
  connect: 'Sambungkan',
  error: 'Gagal Masuk',
  empty_displayname: 'Nama tampilan tidak boleh kosong.',
}

export const context = {
  ignore: 'Abaikan',
  unignore: 'Jangan Abaikan',
  mute: 'Senyapkan',
  unmute: 'Bunyikan',
  release: 'Paksa Lepas Kendali',
  take: 'Paksa Ambil Kendali',
  give: 'Berikan Kendali',
  kick: 'Keluarkan',
  ban: 'Blokir IP',
  confirm: {
    kick_title: 'Keluarkan {name}?',
    kick_text: 'Apakah Anda yakin ingin mengeluarkan {name}?',
    ban_title: 'Blokir {name}?',
    ban_text: 'Apakah Anda ingin memblokir {name}? Anda perlu memulai ulang server untuk membatalkan ini.',
    mute_title: 'Senyapkan {name}?',
    mute_text: 'Apakah Anda yakin ingin mematikan suara {name}?',
    unmute_title: 'Bunyikan {name}?',
    unmute_text: 'Apakah Anda yakin ingin menyalakan suara {name}?',
    button_yes: 'Ya',
    button_cancel: 'Batal',
  },
}

export const controls = {
  release: 'Lepaskan Kendali',
  request: 'Minta Kendali',
  lock: 'Kunci Kendali',
  unlock: 'Buka Kunci Kendali',
  has: 'Anda memiliki kendali',
  hasnot: 'Anda tidak memiliki kendali',
}

export const locks = {
  control: {
    lock: 'Kunci Kendali (untuk pengguna)',
    unlock: 'Buka Kunci Kendali (untuk pengguna)',
    locked: 'Kendali Terkunci (untuk pengguna)',
    unlocked: 'Kendali Terbuka (untuk pengguna)',
    notif_locked: 'mengunci kendali untuk pengguna',
    notif_unlocked: 'membuka kendali untuk pengguna',
  },
  login: {
    lock: 'Kunci Ruangan (untuk pengguna)',
    unlock: 'Buka Kunci Ruangan (untuk pengguna)',
    locked: 'Ruangan Terkunci (untuk pengguna)',
    unlocked: 'Ruangan Terbuka (untuk pengguna)',
    notif_locked: 'mengunci ruangan',
    notif_unlocked: 'membuka ruangan',
  },
  file_transfer: {
    lock: 'Kunci Transfer Berkas (untuk pengguna)',
    unlock: 'Buka Kunci Transfer Berkas (untuk pengguna)',
    locked: 'Transfer Berkas Terkunci (untuk pengguna)',
    unlocked: 'Transfer Berkas Terbuka (untuk pengguna)',
    notif_locked: 'mengunci transfer berkas',
    notif_unlocked: 'membuka transfer berkas',
  },
}

export const setting = {
  scroll: 'Sensitivitas Gulir',
  scroll_invert: 'Gulir Terbalik',
  autoplay: 'Putar Video Otomatis',
  ignore_emotes: 'Abaikan Emoticon',
  chat_sound: 'Nyalakan Bunyi Obrolan',
  keyboard_layout: 'Tata Letak Papan Tik',
  broadcast_title: 'Siaran Langsung',
}

export const connection = {
  logged_out: 'Anda telah keluar.',
  reconnecting: 'Menyambungkan ulang...',
  connected: 'Tersambung',
  disconnected: 'Terputus',
  kicked: 'Anda telah dikeluarkan dari ruangan ini.',
  button_confirm: 'Oke',
}

export const notifications = {
  connected: '{name} tersambung',
  disconnected: '{name} terputus',
  controls_taken: '{name} mengambil kendali',
  controls_taken_force: 'mengambil kendali secara paksa',
  controls_taken_steal: 'mengambil kendali dari {name}',
  controls_released: '{name} melepaskan kendali',
  controls_released_force: 'melepaskan kendali secara paksa',
  controls_released_steal: 'melepaskan kendali dari {name}',
  controls_given: 'memberikan kendali kepada {name}',
  controls_has: '{name} memiliki kendali',
  controls_has_alt: 'Tapi saya sudah memberitahu orang itu bahwa Anda menginginkannya',
  controls_requesting: '{name} meminta kendali',
  resolution: 'mengubah resolusi menjadi {width}x{height}@{rate}',
  banned: 'memblokir {name}',
  kicked: 'mengeluarkan {name}',
  muted: 'mensenyapkan suara {name}',
  unmuted: 'membunyikan suara {name}',
}

export const files = {
  downloads: 'Unduhan',
  uploads: 'Unggahan',
  upload_here: 'Klik atau seret berkas ke sini untuk mengunggah',
}
