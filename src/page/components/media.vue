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

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import type Neko from '@/component/main.vue'

const props = defineProps<{
  neko: typeof Neko
}>()

const audio = ref<HTMLAudioElement | null>(null)
const video = ref<HTMLVideoElement | null>(null)

const audioDevice = ref<string>('')
const audioDevices = ref<MediaDeviceInfo[]>([])

const micTracks = ref<MediaStreamTrack[]>([])
const micSenders = ref<RTCRtpSender[]>([])

const videoDevice = ref<string>('')
const videoDevices = ref<MediaDeviceInfo[]>([])

const camTracks = ref<MediaStreamTrack[]>([])
const camSenders = ref<RTCRtpSender[]>([])

onMounted(() => {
  loadAudioDevices()
  loadVideoDevices()
})

async function loadAudioDevices() {
  let devices = await navigator.mediaDevices.enumerateDevices()
  audioDevices.value = devices.filter((device) => device.kind === 'audioinput')
  console.log('audioDevices', audioDevices.value)
}

async function addMicrophone() {
  micTracks.value = []
  micSenders.value = []

  try {
    let a = { echoCancellation: true } as MediaTrackConstraints
    if (audioDevice.value != '') {
      a.deviceId = audioDevice.value
    }

    const stream = await navigator.mediaDevices.getUserMedia({ video: false, audio: a })
    audio.value!.srcObject = stream
    console.log('Got MediaStream:', stream)

    const tracks = stream.getTracks()
    console.log('Got tracks:', tracks)

    tracks.forEach((track) => {
      micTracks.value.push(track)
      console.log('Adding track', track, stream)

      const rtcp = props.neko.addTrack(track, stream)
      micSenders.value.push(rtcp)
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

function stopMicrophone() {
  micTracks.value.forEach((track) => {
    track.stop()
  })

  micSenders.value.forEach((rtcp) => {
    props.neko.removeTrack(rtcp)
  })

  audio.value!.srcObject = null
  micTracks.value = []
  micSenders.value = []
}

async function loadVideoDevices() {
  let devices = await navigator.mediaDevices.enumerateDevices()
  videoDevices.value = devices.filter((device) => device.kind === 'videoinput')
  console.log('videoDevices', videoDevices.value)
}

async function addWebcam() {
  camTracks.value = []
  camSenders.value = []

  try {
    let v = {
      width: 1280,
      height: 720,
    } as MediaTrackConstraints
    if (videoDevice.value != '') {
      v.deviceId = videoDevice.value
    }

    const stream = await navigator.mediaDevices.getUserMedia({ video: v, audio: false })
    video.value!.srcObject = stream
    console.log('Got MediaStream:', stream)

    const tracks = stream.getTracks()
    console.log('Got tracks:', tracks)

    tracks.forEach((track) => {
      camTracks.value.push(track)
      console.log('Adding track', track, stream)

      const rtcp = props.neko.addTrack(track, stream)
      camSenders.value.push(rtcp)
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

function stopWebcam() {
  camTracks.value.forEach((track) => {
    track.stop()
  })

  camSenders.value.forEach((rtcp) => {
    props.neko.removeTrack(rtcp)
  })

  video.value!.srcObject = null
  camTracks.value = []
  camSenders.value = []
}
</script>
