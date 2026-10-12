<template>
  <video
    ref="video"
    playsinline
    @canplaythrough="$emit('ready')"
    @ended="$emit('ended')"
    @error="$emit('error', $event)"
    @volumechange="onVolumeChange"
    @playing="$emit('playing')"
    @pause="$emit('pause')"
  />
</template>

<script lang="ts">
  import { Component, Prop, Ref, Vue, Watch } from 'vue-property-decorator'

  @Component({ name: 'neko-video-webrtc' })
  export default class VideoWebRTC extends Vue {
    @Ref('video') readonly element!: HTMLVideoElement

    @Prop() readonly stream?: MediaStream
    @Prop(Boolean) readonly playing!: boolean
    @Prop(Number) readonly volume!: number
    @Prop(Boolean) readonly muted!: boolean

    mounted() {
      this.onStreamChanged(this.stream)
      this.onVolumeChanged(this.volume)
      this.onMutedChanged(this.muted)
      void this.onPlayingChanged(this.playing)
    }

    @Watch('stream')
    onStreamChanged(stream?: MediaStream) {
      if (!this.element || !stream) return

      if ('srcObject' in this.element) {
        this.element.srcObject = stream
      } else {
        // @ts-ignore older browsers use a blob URL instead of srcObject
        this.element.src = window.URL.createObjectURL(stream)
      }
    }

    @Watch('volume')
    onVolumeChanged(volume: number) {
      if (this.element && this.element.volume !== volume / 100) this.element.volume = volume / 100
    }

    @Watch('muted')
    onMutedChanged(muted: boolean) {
      if (this.element && this.element.muted !== muted) this.element.muted = muted
    }

    @Watch('playing')
    async onPlayingChanged(playing: boolean) {
      if (!this.element) return
      if (!playing) {
        if (!this.element.paused) this.element.pause()
        return
      }
      if (!this.element.paused) return

      try {
        await this.element.play()
      } catch (_) {
        if (!this.element.muted) {
          try {
            this.element.muted = true
            this.onVolumeChange()
            await this.element.play()
            return
          } catch (_) {}
        }
        this.$emit('play-failed')
      }
    }

    async play() {
      if (this.element.paused) await this.element.play()
    }

    pause() {
      if (!this.element.paused) this.element.pause()
    }

    requestPictureInPicture() {
      // @ts-ignore requestPictureInPicture is not present in every DOM type definition
      return this.element.requestPictureInPicture()
    }

    private onVolumeChange() {
      this.$emit('volume-change', { muted: this.element.muted, volume: this.element.volume * 100 })
    }
  }
</script>
