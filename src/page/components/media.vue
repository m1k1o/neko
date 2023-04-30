<template>
  <div class="media" style="width: 100%">
    <button v-if="micTracks.length == 0" @click="addMicrophone">Add microphone</button>
    <button v-else @click="stopMicrophone">Stop microphone</button>
    <select v-model="audioDevice">
      <option value="">Default microphone</option>
      <option v-for="device in audioDevices" :key="device.deviceId" :value="device.deviceId">
        {{ device.label || 'Unnamed input' }}
      </option>
    </select>
    <i
      style="margin: 0 5px; cursor: pointer"
      class="fa fa-refresh"
      title="Reload audio devices"
      @click="loadAudioDevices"
    ></i>
    <br />
    <audio v-show="micTracks.length > 0" ref="audio" controls />
    <hr />
    <button v-if="camTracks.length == 0" @click="addWebcam">Add webcam</button>
    <button v-else @click="stopWebcam">Stop webcam</button>
    <select v-model="videoDevice">
      <option value="">Default webcam</option>
      <option v-for="device in videoDevices" :key="device.deviceId" :value="device.deviceId">
        {{ device.label || 'Unnamed input' }}
      </option>
    </select>
    <i
      style="margin: 0 5px; cursor: pointer"
      class="fa fa-refresh"
      title="Reload video devices"
      @click="loadVideoDevices"
    ></i>
    <br />
    <video v-show="camTracks.length > 0" ref="video" controls />
    <br />
    <p>Video must be enabled and supported by the server.</p>
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

    private audioDevice: string = ''
    private audioDevices: MediaDeviceInfo[] = []

    private micTracks: MediaStreamTrack[] = []
    private micSenders: RTCRtpSender[] = []

    private videoDevice: string = ''
    private videoDevices: MediaDeviceInfo[] = []

    private camTracks: MediaStreamTrack[] = []
    private camSenders: RTCRtpSender[] = []

    mounted() {
      this.loadAudioDevices()
      this.loadVideoDevices()
    }

    async loadAudioDevices() {
      let devices = await navigator.mediaDevices.enumerateDevices()
      this.audioDevices = devices.filter((device) => device.kind === 'audioinput')
      console.log('audioDevices', this.audioDevices)
    }

    async addMicrophone() {
      this.micTracks = []
      this.micSenders = []

      try {
        let audio = { echoCancellation: true } as MediaTrackConstraints
        if (this.audioDevice != '') {
          audio.deviceId = this.audioDevice
        }

        const stream = await navigator.mediaDevices.getUserMedia({ video: false, audio })
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

    async loadVideoDevices() {
      let devices = await navigator.mediaDevices.enumerateDevices()
      this.videoDevices = devices.filter((device) => device.kind === 'videoinput')
      console.log('videoDevices', this.videoDevices)
    }

    async addWebcam() {
      this.camTracks = []
      this.camSenders = []

      try {
        let video = {
          width: 1280,
          height: 720,
        } as MediaTrackConstraints
        if (this.videoDevice != '') {
          video.deviceId = this.videoDevice
        }

        const stream = await navigator.mediaDevices.getUserMedia({ video, audio: false })
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
  }
</script>
