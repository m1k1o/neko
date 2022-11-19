export const logout = '로그아웃'
export const unsupported = '이 브라우저는 WebRTC 를 지원하지 않습니다'
export const admin_loggedin = '관리자로 로그인했습니다'
export const you = '당신'
export const somebody = '누군가'
export const send_a_message = '메세지 보내기'

export const side = {
  chat: '채팅',
  files: '파일',
  settings: '설정',
}

export const connect = {
  login_title: '로그인 해주세요',
  invitation_title: '이 방에 초대됐습니다',
  displayname: '표시될 이름을 입력해주세요',
  password: '비밀번호',
  connect: '연결',
  error: '로그인 오류',
  empty_displayname: '표시될 이름은 비어있을 수 없습니다.',
}

export const context = {
  ignore: '무시하기',
  unignore: '무시하기 해제',
  mute: '뮤트',
  unmute: '뮤트 해제',
  release: '강제로 조작 권한 풀기',
  take: '강제로 조작 권한 가져오기',
  give: '조작 권한 주기',
  kick: '추방',
  ban: '아이피 차단',
  confirm: {
    kick_title: '{name}님을 추방할까요?',
    kick_text: '정말로 {name}님을 추방할까요?',
    ban_title: '{name}님을 차단할까요?',
    ban_text: '정말로 {name}님을 차단할까요? 이 작업을 되돌릴려면 서버를 다시 시작해야합니다.',
    mute_title: '{name}님을 뮤트할까요?',
    mute_text: '정말로 {name}님을 뮤트할까요?',
    unmute_title: '{name}님의 뮤트를 해제할까요?',
    unmute_text: '정말로 {name}님의 뮤트를 해제할까요?',
    button_yes: '네',
    button_cancel: '아니요',
  },
}

export const controls = {
  release: '조작 권한 풀기',
  request: '조작 권한 요청',
  lock: '조작 잠그기',
  unlock: '조작 잠금 해제하기',
}

export const locks = {
  control: {
    lock: '조작 잠그기 (사용자)',
    unlock: '조작 잠금 해제하기 (사용자)',
    locked: '조작이 잠겼습니다 (사용자)',
    unlocked: '조작 잠금이 해제됐습니다 (사용자)',
    notif_locked: '사용자의 조작을 잠궜습니다',
    notif_unlocked: '사용자의 조작 잠금을 해제했습니다',
  },
  login: {
    lock: '방 잠그기 (사용자)',
    unlock: '방 잠금 해제하기 (사용자)',
    locked: '방이 잠겼습니다 (사용자)',
    unlocked: '방 잠금이 해제됐습니다 (사용자)',
    notif_locked: '방이 잠겼습니다',
    notif_unlocked: '방 잠금이 해제됐습니다',
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
  scroll: '스크롤 감도',
  scroll_invert: '스크롤 반전',
  autoplay: '동영상 자동 재생',
  ignore_emotes: '이모지 무시',
  chat_sound: '채팅 소리 재생',
  keyboard_layout: '키보드 레이아웃',
  broadcast_title: '실시간 방송',
}

export const connection = {
  logged_out: '로그아웃 했습니다.',
  reconnecting: '다시 접속하는 중...',
  connected: '연결됨',
  disconnected: '연결 해제됨',
  kicked: '이 방에서 추방됐습니다.',
  button_confirm: '확인',
}

export const notifications = {
  connected: '{name} 님이 접속하셨습니다',
  disconnected: '{name} 님이 퇴장하셨습니다',
  controls_taken: '{name} 님이 조작 권한을 가지셨습니다',
  controls_taken_force: '조작 권한을 강제로 가졌습니다',
  controls_taken_steal: '{name} 님으로 부터 조작 권한을 가져왔습니다',
  controls_released: '{name} 님이 조작 권한을 내려놨습니다',
  controls_released_force: '조작 권한을 강제로 내려놨습니다',
  controls_released_steal: '{name} 님의 조작 권한을 내려놨습니다',
  controls_given: '{name} 님에게 조작 권한을 줬습니다',
  controls_has: '{name} 님이 조작 권한을 가지고 있습니다',
  controls_has_alt: '가지고 있는 사람에게 가지고 싶어한다고 알려드리겠습니다',
  controls_requesting: '{name}님이 조작 권한을 요청하셨습니다',
  resolution: '해상도로 {width}x{height}@{rate} 로 변경했습니다',
  banned: '{name} 님이 차단됐습니다',
  kicked: '{name} 님이 추방됐습니다',
  muted: '{name} 님이 뮤트됐습니다',
  unmuted: '{name} 님의 뮤트가 해제됐습니다',
}

export const files = {
  downloads: '다운로드',
  uploads: '업로드',
  upload_here: '업로드할 파일을 여기로 클릭하거나 드래그하세요.',
}
