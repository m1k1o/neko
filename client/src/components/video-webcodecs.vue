<template>
  <canvas ref="canvas" />
</template>

<script lang="ts">
  import { Component, Prop, Ref, Vue, Watch } from 'vue-property-decorator'
  import { MediaConfiguration } from '~/neko/messages'
  import { WebCodecsPlayer } from '~/neko/webcodecs'

  @Component({ name: 'neko-video-webcodecs' })
  export default class VideoWebCodecs extends Vue {
    @Ref('canvas') readonly element!: HTMLCanvasElement

    @Prop(Object) readonly media?: MediaConfiguration
    @Prop(Boolean) readonly playing!: boolean
    @Prop(Number) readonly volume!: number
    @Prop(Boolean) readonly muted!: boolean

    private player?: WebCodecsPlayer
    private generation = 0

    mounted() {
      void this.onMediaChanged(this.media)
    }

    beforeDestroy() {
      this.generation++
      this.stop()
    }

    @Watch('media')
    async onMediaChanged(media?: MediaConfiguration) {
      const generation = ++this.generation
      this.stop()
      if (!media) return

      const player = new WebCodecsPlayer(
        this.element,
        media,
        () => {
          if (this.player === player) this.$emit('ready')
        },
        (error) => {
          if (this.player !== player) return
          this.player = undefined
          this.$emit('error', error)
        },
      )
      this.player = player

      try {
        player.setVolume(this.volume / 100)
        player.setMuted(this.muted)
        await player.setPlaying(this.playing)
        await player.start()
        if (generation !== this.generation) player.stop()
      } catch (error) {
        player.stop()
        if (this.player === player) this.player = undefined
        if (generation === this.generation) this.$emit('error', error)
      }
    }

    @Watch('playing')
    async onPlayingChanged(playing: boolean) {
      try {
        await this.player?.setPlaying(playing)
      } catch (error) {
        this.$emit('error', error)
      }
    }

    @Watch('volume')
    onVolumeChanged(volume: number) {
      this.player?.setVolume(volume / 100)
    }

    @Watch('muted')
    onMutedChanged(muted: boolean) {
      this.player?.setMuted(muted)
    }

    private stop() {
      this.player?.stop()
      this.player = undefined
    }
  }
</script>
