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

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'

  @Component({
    name: 'neko-color',
  })
  export default class extends Vue {
    @Ref('canvas') readonly _canvas!: HTMLCanvasElement
    @Ref('color') readonly _color!: HTMLDivElement

    ctx!: CanvasRenderingContext2D | null
    video!: HTMLVideoElement | null
    interval!: number
    picker!: HTMLDivElement | null
    bullet!: HTMLDivElement | null
    color!: string

    x = 0
    y = 0
    clickOnChange = false

    mounted() {
      this.video = document.querySelector('video')
      this.ctx = this._canvas.getContext('2d')
    }

    beforeDestroy() {
      if (this.interval) {
        window.clearInterval(this.interval)
      }

      this.clearPoint()
    }

    @Watch('clickOnChange')
    clickOnChangeChanged() {
      if (this.clickOnChange) {
        // register interval timer
        this.interval = window.setInterval(this.intervalTimer, 0)
      } else {
        // unregister interval timer
        window.clearInterval(this.interval)
        this.color = ''
      }
    }

    intervalTimer() {
      if (!this.video || !this.ctx) {
        return
      }

      this._canvas.width = this.video.videoWidth
      this._canvas.height = this.video.videoHeight
      this.ctx.clearRect(0, 0, this.video.videoWidth, this.video.videoHeight)
      this.ctx.drawImage(this.video, 0, 0, this.video.videoWidth, this.video.videoHeight)

      // get color from pixel at x,y
      var pixel = this.ctx.getImageData(this.x, this.y, 1, 1)
      var data = pixel.data
      var rgba = 'rgba(' + data[0] + ', ' + data[1] + ', ' + data[2] + ', ' + data[3] / 255 + ')'

      // if color is different, update it
      if (this.color != rgba) {
        if (this.clickOnChange && this.color) {
          this.$emit('colorChange', { x: this.x, y: this.y })
          this.clickOnChange = false
        }

        console.log('color change', rgba, this.color)
        this._color.style.backgroundColor = rgba
        this.color = rgba
      }
    }

    getCoords(elem: HTMLElement) {
      // crossbrowser version
      let box = elem.getBoundingClientRect()

      let body = document.body
      let docEl = document.documentElement

      let scrollTop = window.pageYOffset || docEl.scrollTop || body.scrollTop
      let scrollLeft = window.pageXOffset || docEl.scrollLeft || body.scrollLeft

      let clientTop = docEl.clientTop || body.clientTop || 0
      let clientLeft = docEl.clientLeft || body.clientLeft || 0

      let top = box.top + scrollTop - clientTop
      let left = box.left + scrollLeft - clientLeft

      return { top: Math.round(top), left: Math.round(left) }
    }

    setPoint() {
      // create new element and add to body
      var picker = document.createElement('div')

      // coordinates of video element
      var video = this.getCoords(this.video!)

      // match that dimensions and offset matches video
      picker.style.width = this.video!.offsetWidth + 'px'
      picker.style.height = this.video!.offsetHeight + 'px'
      picker.style.left = video.left + 'px'
      picker.style.top = video.top + 'px'
      picker.style.position = 'absolute'
      picker.style.backgroundColor = 'rgba(0, 0, 0, 0.5)'
      picker.style.cursor = 'crosshair'

      // put it on top of video
      picker.style.zIndex = '100'

      document.body.appendChild(picker)

      // add click event listener to new element
      picker.addEventListener('click', this.clickPicker)
      this.picker = picker
    }

    clearPoint() {
      this.x = 0
      this.y = 0
      this.color = ''
      this._color.style.backgroundColor = 'transparent'

      if (this.bullet) {
        this.bullet.remove()
      }

      if (this.picker) {
        this.picker.remove()
      }
    }

    clickPicker(e: any) {
      // get mouse position
      var x = e.pageX
      var y = e.pageY

      // get picker position
      var picker = this.getCoords(this.picker!)

      // calculate new x,y position
      var newX = x - picker.left
      var newY = y - picker.top

      // make it relative to video size
      newX = Math.round((newX / this.video!.offsetWidth) * this.video!.videoWidth)
      newY = Math.round((newY / this.video!.offsetHeight) * this.video!.videoHeight)

      console.log(newX, newY)

      // set new x,y position
      this.x = newX
      this.y = newY

      // remove picker element
      this.picker!.remove()

      // add bullet element to the position
      if (this.bullet) {
        this.bullet.remove()
      }
      var bullet = document.createElement('div')
      bullet.style.left = x + 'px'
      bullet.style.top = y + 'px'
      // width and height of bullet
      bullet.style.width = '10px'
      bullet.style.height = '10px'
      // background color of bullet
      bullet.style.backgroundColor = 'red'
      // border radius of bullet
      bullet.style.borderRadius = '50%'
      // transform bullet to center
      bullet.style.transform = 'translate(-50%, -50%)'
      bullet.style.position = 'absolute'
      bullet.style.zIndex = '100'
      document.body.appendChild(bullet)
      this.bullet = bullet
    }
  }
</script>
