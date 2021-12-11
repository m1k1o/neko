export function makeid(length: number) {
  let result = ''
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const charactersLength = characters.length
  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength))
  }
  return result
}

export function elementRequestFullscreen(el: HTMLElement) {
  if (typeof el.requestFullscreen === 'function') {
    el.requestFullscreen()
    //@ts-ignore
  } else if (typeof el.webkitRequestFullscreen === 'function') {
    //@ts-ignore
    el.webkitRequestFullscreen()
    //@ts-ignore
  } else if (typeof el.webkitEnterFullscreen === 'function') {
    //@ts-ignore
    el.webkitEnterFullscreen()
    //@ts-ignore
  } else if (typeof el.msRequestFullScreen === 'function') {
    //@ts-ignore
    el.msRequestFullScreen()
  } else {
    return false
  }
  return true
}
