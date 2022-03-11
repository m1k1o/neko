// Takes a snapshot of the video and returns as object url
export function videoSnap(video: HTMLVideoElement): Promise<string> {
  return new Promise((res, rej) => {
    const width = video.videoWidth
    const height = video.videoHeight

    const canvas = document.createElement('canvas') as HTMLCanvasElement
    canvas.width = width
    canvas.height = height

    const ctx = canvas.getContext('2d')
    if (ctx === null) {
      rej('canvas context is null')
      return
    }

    // Define the size of the rectangle that will be filled (basically the entire element)
    ctx.fillRect(0, 0, width, height)

    // Grab the image from the video
    ctx.drawImage(video, 0, 0, width, height)

    canvas.toBlob(function (blob) {
      if (blob === null) {
        rej('canvas blob is null')
        return
      }

      const url = URL.createObjectURL(blob)
      res(url)
    })
  })
}
