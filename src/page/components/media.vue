<template>
  <div class="media" style="width: 100%">
    <!--
    <button @click="getDevices">List available devices</button>
    <button v-for="d in devices" :key="d.deviceId">
      {{ d.kind }} : {{ d.label }} id = {{ d.deviceId }}
    </button>
    -->
    <button v-if="micTracks.length == 0" @click="addMicrophone">Add microphone</button>
    <button v-else @click="stopMicrophone">Stop microphone</button>
    <br />
    <audio v-show="micTracks.length > 0" ref="audio" controls />
    <hr />
    <button v-if="camTracks.length == 0" @click="addWebcam">Add webcam</button>
    <button v-else @click="stopWebcam">Stop webcam</button>
    <br />
    <video v-show="camTracks.length > 0" ref="video" controls />
  </div>
</template>

<style lang="scss" scoped></style>

<script lang="ts">
  import { Vue, Component, Prop, Ref } from 'vue-property-decorator'
  import Neko from '~/component/main.vue'

  @Component({
    name: 'neko-media',
  })
  export default class extends Vue {
    @Prop() readonly neko!: Neko
    @Ref('audio') readonly _audio!: HTMLAudioElement
    @Ref('video') readonly _video!: HTMLVideoElement

    private micTracks: MediaStreamTrack[] = []
    private micSenders: RTCRtpSender[] = []

    private camTracks: MediaStreamTrack[] = []
    private camSenders: RTCRtpSender[] = []

    //private devices: any[] = []

    async addMicrophone() {
      this.micTracks = []
      this.micSenders = []

      try {
        const stream = await navigator.mediaDevices.getUserMedia({ video: false, audio: true })

        this._audio.srcObject = stream
        console.log('Got MediaStream:', stream)

        const tracks = stream.getTracks()
        console.log('Got tracks:', tracks)

        tracks.forEach((track) => {
          this.micTracks.push(track)
          console.log('Adding track', track, stream)

          const rtcp = this.neko.addTrack(track, stream)
          this.micSenders.push(rtcp)
          console.log('rtcp sender', rtcp, rtcp.transport)

          // TODO: Can be null.
          rtcp.transport?.addEventListener('statechange', () => {
            console.log('track - on state change', rtcp.transport?.state)
          })
        })
      } catch (error) {
        alert('Error accessing media devices.' + error)
      }
    }

    stopMicrophone() {
      this.micTracks.forEach((track) => {
        track.stop()
      })

      this.micSenders.forEach((rtcp) => {
        this.neko.removeTrack(rtcp)
      })

      this._audio.srcObject = null
      this.micTracks = []
      this.micSenders = []
    }

    async addWebcam() {
      this.camTracks = []
      this.camSenders = []

      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: {
            width: 1280,
            height: 720,
          },
          audio: false,
        })

        this._video.srcObject = stream
        console.log('Got MediaStream:', stream)

        const tracks = stream.getTracks()
        console.log('Got tracks:', tracks)

        tracks.forEach((track) => {
          this.camTracks.push(track)
          console.log('Adding track', track, stream)

          const rtcp = this.neko.addTrack(track, stream)
          this.camSenders.push(rtcp)
          console.log('rtcp sender', rtcp, rtcp.transport)

          // TODO: Can be null.
          rtcp.transport?.addEventListener('statechange', () => {
            console.log('track - on state change', rtcp.transport?.state)
          })
        })
      } catch (error) {
        alert('Error accessing media devices.' + error)
      }
    }

    stopWebcam() {
      this.camTracks.forEach((track) => {
        track.stop()
      })

      this.camSenders.forEach((rtcp) => {
        this.neko.removeTrack(rtcp)
      })

      this._audio.srcObject = null
      this.camTracks = []
      this.camSenders = []
    }

    /*async getDevices() {
      const devices = await navigator.mediaDevices.enumerateDevices();
      this.devices = devices.map(({ kind, label, deviceId }) => ({ kind, label, deviceId }))
      console.log(this.devices)
    }*/
  }
</script>
