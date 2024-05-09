<template>
  <div class="color">
    <canvas ref="canvas"></canvas>
    <div class="color-box" ref="color"></div>
    <div v-if="x && y">Point: ({{ x }}, {{ y }})</div>
    <div v-else>No point set</div>
    <button @click="setPoint">Set</button>
    <button @click="clearPoint">Clear</button>
    <br />
    <button v-if="!clickOnChange" @click="clickOnChange = true">Click on color change</button>
    <button v-if="clickOnChange" @click="clickOnChange = false">Don't click</button>
  </div>
</template>

<style lang="scss" scoped>
  @import '@/page/assets/styles/main.scss';

  canvas {
    width: 1px;
    height: 1px;
    position: absolute;
  }

  .color {
    width: 100%;
    height: 100%;
    text-align: center;
  }

  .color-box {
    display: inline-block;
    width: 100%;
    height: 10px;
    margin-bottom: 5px;
  }
</style>

<script lang="ts" setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'

const canvas = ref<HTMLCanvasElement | null>(null)
const color = ref<HTMLDivElement | null>(null)

const ctx = ref<CanvasRenderingContext2D | null>(null)
const video = ref<HTMLVideoElement | null>(null)
const interval = ref<number>(0)
const picker = ref<HTMLDivElement | null>(null)
const bullet = ref<HTMLDivElement | null>(null)
const currColor = ref<string>('')

const x = ref<number>(0)
const y = ref<number>(0)
const clickOnChange = ref<boolean>(false)

const emit = defineEmits(['colorChange'])

onMounted(() => {
  video.value = document.querySelector('video')
  ctx.value = canvas.value?.getContext('2d') || null
})

onBeforeUnmount(() => {
  if (interval.value) {
    window.clearInterval(interval.value)
  }

  clearPoint()
})

function clickOnChangeChanged() {
  if (clickOnChange.value) {
    // register interval timer
    interval.value = window.setInterval(intervalTimer, 0)
  } else {
    // unregister interval timer
    window.clearInterval(interval.value)
    currColor.value = ''
  }
}

watch(clickOnChange, clickOnChangeChanged)

function intervalTimer() {
  if (!video.value || !ctx.value) {
    return
  }

  canvas.value!.width = video.value.videoWidth
  canvas.value!.height = video.value.videoHeight
  ctx.value.clearRect(0, 0, video.value.videoWidth, video.value.videoHeight)
  ctx.value.drawImage(video.value, 0, 0, video.value.videoWidth, video.value.videoHeight)

  // get color from pixel at x,y
  const pixel = ctx.value.getImageData(x.value, y.value, 1, 1)
  const data = pixel.data
  const rgba = `rgba(${data[0]}, ${data[1]}, ${data[2]}, ${data[3] / 255})`

  // if color is different, update it
  if (currColor.value !== rgba) {
    if (clickOnChange.value && currColor.value) {
      emit('colorChange', { x: x.value, y: y.value })
      clickOnChange.value = false
    }

    console.log('color change', rgba, currColor.value)
    color.value!.style.backgroundColor = rgba
    currColor.value = rgba
  }
}

function getCoords(elem: HTMLElement) {
  // crossbrowser version
  const box = elem.getBoundingClientRect()

  const body = document.body
  const docEl = document.documentElement

  const scrollTop = window.pageYOffset || docEl.scrollTop || body.scrollTop
  const scrollLeft = window.pageXOffset || docEl.scrollLeft || body.scrollLeft

  const clientTop = docEl.clientTop || body.clientTop || 0
  const clientLeft = docEl.clientLeft || body.clientLeft || 0

  const top = box.top + scrollTop - clientTop
  const left = box.left + scrollLeft - clientLeft

  return { top: Math.round(top), left: Math.round(left) }
}

function setPoint() {
  // create new element and add to body
  const p = document.createElement('div')

  // coordinates of video element
  const v = getCoords(video.value!)

  // match that dimensions and offset matches video
  p.style.width = video.value!.offsetWidth + 'px'
  p.style.height = video.value!.offsetHeight + 'px'
  p.style.left = v.left + 'px'
  p.style.top = v.top + 'px'
  p.style.position = 'absolute'
  p.style.backgroundColor = 'rgba(0, 0, 0, 0.5)'
  p.style.cursor = 'crosshair'

  // put it on top of video
  p.style.zIndex = '100'

  document.body.appendChild(p)

  // add click event listener to new element
  p.addEventListener('click', clickPicker)
  picker.value = p
}

function clearPoint() {
  x.value = 0
  y.value = 0
  color.value!.style.backgroundColor = 'transparent'

  if (bullet.value) {
    bullet.value.remove()
  }

  if (picker.value) {
    picker.value.remove()
  }
}

function clickPicker(e: any) {
  // get picker position
  const p = getCoords(picker.value!)

  // get mouse position
  let pageX = e.pageX
  let pageY = e.pageY

  // make it relative to video size and save it
  x.value = Math.round(((pageX - p.left) / video.value!.offsetWidth) * video.value!.videoWidth)
  y.value = Math.round(((pageY - p.top) / video.value!.offsetHeight) * video.value!.videoHeight)

  // remove picker element
  picker.value!.remove()

  // add bullet element to the position
  if (bullet.value) {
    bullet.value.remove()
  }
  const b = document.createElement('div')
  b.style.left = pageX + 'px'
  b.style.top = pageY + 'px'
  // width and height of bullet
  b.style.width = '10px'
  b.style.height = '10px'
  // background color of bullet
  b.style.backgroundColor = 'red'
  // border radius of bullet
  b.style.borderRadius = '50%'
  // transform bullet to center
  b.style.transform = 'translate(-50%, -50%)'
  b.style.position = 'absolute'
  b.style.zIndex = '100'
  document.body.appendChild(b)
  bullet.value = b
}
</script>
