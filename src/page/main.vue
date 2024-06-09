<template>
  <div id="neko" :class="[expanded ? 'expanded' : '']">
    <main class="neko-main">
      <div class="header-container" v-if="neko">
        <NekoHeader :neko="neko" @toggle="expanded = !expanded" />
      </div>
      <div class="video-container">
        <NekoCanvas ref="neko" :server="server" autologin autoconnect autoplay />
        <div v-if="loaded && neko!.private_mode_enabled" class="player-notif">Private mode is currently enabled.</div>
        <div
          v-if="loaded && neko!.state.connection.type === 'webrtc' && !neko!.state.video.playing"
          class="player-overlay"
        >
          <i @click.stop.prevent="neko!.play()" v-if="neko!.state.video.playable" class="fas fa-play-circle" />
        </div>
        <div v-if="uploadActive" class="player-overlay" style="background: rgba(0, 0, 0, 0.8); font-size: 1vw">
          UPLOAD IN PROGRESS: {{ Math.round(uploadProgress) }}%
        </div>
        <div
          v-else-if="dialogOverlayActive"
          class="player-overlay"
          style="background: rgba(0, 0, 0, 0.8); font-size: 1vw"
        >
          SOMEONE IS UPLOADING A FILE, PLEASE WAIT
        </div>
        <div
          v-else-if="dialogRequestActive"
          class="player-overlay"
          style="background: rgba(0, 0, 0, 0.8); font-size: 1vw; flex-flow: column"
          @dragenter.stop.prevent
          @dragleave.stop.prevent
          @dragover.stop.prevent
          @drop.stop.prevent="dialogUploadFiles([...$event.dataTransfer!.files])"
        >
          <span style="padding: 1em">UPLOAD REQUESTED:</span>
          <span style="background: white">
            <input type="file" @change="dialogUploadFiles([...($event.target as HTMLInputElement)!.files!])" multiple />
          </span>
          <span style="padding: 1em; padding-bottom: 0; font-style: italic">(or drop files here)</span>
          <span style="padding: 1em">
            <button @click="dialogCancel()">CANCEL</button>
          </span>
        </div>
      </div>
      <div class="room-container" style="text-align: center">
        <button
          v-if="loaded && isTouchDevice"
          @click="neko!.mobileKeyboardToggle"
          style="position: absolute; left: 5px; transform: translateY(-100%)"
        >
          <i class="fa fa-keyboard" />
        </button>
        <span v-if="loaded && neko!.state.session_id" style="padding-top: 10px">
          You are logged in as
          <strong style="font-weight: bold">
            {{ neko!.state.sessions[neko!.state.session_id].profile.name }}
          </strong>
        </span>

        <div class="room-menu">
          <div class="left-menu">
            <button @click="toggleCursor">
              <i v-if="usesCursor" class="fas fa-mouse-pointer" />
              <i v-else class="fas fa-location-arrow" />
            </button>
          </div>
          <div class="controls">
            <template v-if="loaded && neko">
              <NekoConnect v-if="neko!.state.connection.status == 'disconnected'" :neko="neko" />
              <NekoControls v-else :neko="neko" />
            </template>
          </div>
          <div class="right-menu">
            <div style="text-align: right" v-if="loaded">
              <button v-if="neko!.state.connection.status != 'disconnected'" @click="neko!.disconnect()">
                disconnect
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
    <aside class="neko-menu" v-if="expanded">
      <div class="tabs-container">
        <ul>
          <li :class="{ active: tab === 'events' }" @click.prevent="tab = 'events'">
            <i class="fas fa-sliders-h" />
            <span v-show="tab === 'events'">Events</span>
          </li>
          <li :class="{ active: tab === 'members' }" @click.prevent="tab = 'members'">
            <i class="fas fa-users" />
            <span v-show="tab === 'members'">Members</span>
          </li>
          <li :class="{ active: tab === 'media' }" @click.prevent="tab = 'media'">
            <i class="fas fa-microphone" />
            <span v-show="tab === 'media'">Media</span>
          </li>

          <!-- Plugins -->
          <component v-for="(el, key) in pluginsTabs" :key="key" :is="el" :tab="tab" @tab="tab = $event" />
        </ul>
      </div>
      <div class="page-container" v-if="neko">
        <NekoEvents v-if="tab === 'events'" :neko="neko" />
        <NekoMembers v-if="tab === 'members'" :neko="neko" />
        <NekoMedia v-if="tab === 'media'" :neko="neko" />

        <!-- Plugins -->
        <component v-for="(el, key) in pluginsComponents" :key="key" :is="el" :tab="tab" :neko="neko" />
      </div>
    </aside>
  </div>
</template>

<style lang="scss">
  @import '@/page/assets/styles/main.scss';

  .video-container {
    position: relative;
    overflow: hidden;
    width: 100%;
    height: 100%;
  }

  .player-notif {
    position: absolute;
    top: 0;
    overflow: hidden;
    background: #2a5f2a;
    padding: 10px;
  }

  .player-overlay {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 100%;
    height: 100%;
    overflow: hidden;

    background: rgba($color: #000, $alpha: 0.2);
    display: flex;
    justify-content: center;
    align-items: center;

    i {
      cursor: pointer;
      &::before {
        font-size: 120px;
        text-align: center;
      }
    }

    &.hidden {
      display: none;
    }
  }

  #neko {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    max-width: 100vw;
    max-height: 100vh;
    flex-direction: row;
    display: flex;
  }

  .neko-main {
    min-width: 360px;
    max-width: 100%;
    flex-grow: 1;
    flex-direction: column;
    display: flex;
    overflow: auto;

    .header-container {
      background: $background-tertiary;
      height: $menu-height;
      flex-shrink: 0;
    }

    .video-container {
      background: rgba($color: #000, $alpha: 0.4);
      max-width: 100%;
      flex-grow: 1;
    }

    .room-container {
      background: $background-tertiary;
      height: $controls-height;
      max-width: 100%;
      flex-shrink: 0;
      flex-direction: column;
      display: flex;
      /* for mobile */
      overflow-y: hidden;
      overflow-x: auto;

      .room-menu {
        max-width: 100%;
        flex: 1;
        display: flex;

        .left-menu {
          margin-left: 10px;
          flex: 1;
          justify-content: flex-start;
          align-items: center;
          display: flex;
        }

        .controls {
          flex: 1;
          justify-content: center;
          align-items: center;
          display: flex;
        }

        .right-menu {
          margin-right: 10px;
          flex: 1;
          justify-content: flex-end;
          align-items: center;
          display: flex;
        }
      }
    }
  }

  .neko-menu {
    width: $side-width;
    background-color: $background-primary;
    flex-shrink: 0;
    max-height: 100%;
    max-width: 100%;
    display: flex;
    flex-direction: column;

    .tabs-container {
      background: $background-tertiary;
      height: $menu-height;
      max-height: 100%;
      max-width: 100%;
      display: flex;
      flex-shrink: 0;

      ul {
        display: inline-block;
        padding: 16px 0 0 0;
        overflow-x: auto;

        li {
          background: $background-secondary;
          border-radius: 3px 3px 0 0;
          border-bottom: none;
          display: inline-block;
          padding: 5px 10px;
          margin-right: 4px;
          font-weight: 600;
          cursor: pointer;

          i {
            margin-right: 4px;
            font-size: 10px;
          }

          &.active {
            background: $background-primary;
          }
        }
      }
    }

    .page-container {
      max-height: 100%;
      flex-grow: 1;
      display: flex;
      flex-direction: column;
      overflow: auto;
      padding: 5px;
      box-sizing: border-box;
    }
  }

  /* for mobile */
  @media only screen and (max-width: 600px) {
    $offset: 38px;

    #neko.expanded {
      /* show only enough of the menu to see the toggle button */
      .neko-main {
        transform: translateX(calc(-100% + $offset));
        video {
          display: none;
        }
      }
      .neko-menu {
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        left: $offset;
        width: calc(100% - $offset);
      }
      /* display menu toggle button far right */
      .header .menu,
      .header .menu li {
        margin-right: 2px;
      }
    }
  }
</style>

<script lang="ts" setup>
// plugins must be available at:
// ./plugins/{name}/index.ts -> { Components, Tabs }
const plugins = import.meta.glob('./plugins/*/index.ts')

const pluginsTabs = shallowRef<Record<string, any>>({})
const pluginsComponents = shallowRef<Record<string, any>>({})

// dynamic plugins loader
onMounted(async () => {
  const resolvedPlugins = await Promise.all(
    Object.entries(plugins).map(async ([path, component]) => {
      return [path, await component()]
    }),
  ) as [string, { Components: any, Tabs: any }][]

  pluginsTabs.value = {}
  pluginsComponents.value = {}
  for (const [path, { Components, Tabs }] of resolvedPlugins) {
    pluginsTabs.value[path] = Tabs
    pluginsComponents.value[path] = Components
  }
})

import { ref, shallowRef, computed, onMounted } from 'vue'

import type { AxiosProgressEvent } from 'axios'
import NekoCanvas from '@/component/main.vue'
import type { Settings } from '@/component/types/state'
import NekoHeader from './components/header.vue'
import NekoConnect from './components/connect.vue'
import NekoControls from './components/controls.vue'
import NekoEvents from './components/events.vue'
import NekoMembers from './components/members.vue'
import NekoMedia from './components/media.vue'

const neko = ref<typeof NekoCanvas>()

const expanded = ref(!window.matchMedia('(max-width: 600px)').matches) // default to expanded on bigger screens
const loaded = ref(false)
const tab = ref('')

const server = ref(location.href)

const uploadActive = ref(false)
const uploadProgress = ref(0)

const isTouchDevice = computed(() => 'ontouchstart' in window || navigator.maxTouchPoints > 0)

const dialogOverlayActive = ref(false)
const dialogRequestActive = ref(false)
async function dialogUploadFiles(files: File[]) {
  console.log('will upload files', files)

  uploadActive.value = true
  uploadProgress.value = 0
  try {
    await neko.value!.room.uploadDialog([...files], {
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (!progressEvent.total) {
          uploadProgress.value = 0
          return
        }
        uploadProgress.value = (progressEvent.loaded / progressEvent.total) * 100
      },
    })
  } catch (e: any) {
    alert(e.response ? e.response.data.message : e)
  } finally {
    uploadActive.value = false
  }
}

function dialogCancel() {
  neko.value!.room.uploadDialogClose()
}

onMounted(() => {
  loaded.value = true
  tab.value = 'events'
  //@ts-ignore
  window.neko = neko

  // initial URL
  const url = new URL(location.href).searchParams.get('url')
  if (url) {
    server.value = url
  }

  //
  // connection events
  //
  neko.value!.events.on('connection.status', (status: 'connected' | 'connecting' | 'disconnected') => {
    console.log('connection.status', status)
  })
  neko.value!.events.on('connection.type', (type: 'fallback' | 'webrtc' | 'none') => {
    console.log('connection.type', type)
  })
  neko.value!.events.on('connection.webrtc.sdp', (type: 'local' | 'remote', data: string) => {
    console.log('connection.webrtc.sdp', type, data)
  })
  neko.value!.events.on('connection.webrtc.sdp.candidate', (type: 'local' | 'remote', data: RTCIceCandidateInit) => {
    console.log('connection.webrtc.sdp.candidate', type, data)
  })
  neko.value!.events.on('connection.closed', (error?: Error) => {
    if (error) {
      alert('Connection closed with error: ' + error.message)
    } else {
      alert('Connection closed without error.')
    }
  })

  //
  // drag and drop events
  //
  neko.value!.events.on('upload.drop.started', () => {
    uploadActive.value = true
    uploadProgress.value = 0
  })
  neko.value!.events.on('upload.drop.progress', (progressEvent: AxiosProgressEvent) => {
    if (!progressEvent.total) {
      uploadProgress.value = 0
      return
    }
    uploadProgress.value = (progressEvent.loaded / progressEvent.total) * 100
  })
  neko.value!.events.on('upload.drop.finished', (e?: any) => {
    uploadActive.value = false
    if (e) {
      alert(e.response ? e.response.data.message : e)
    }
  })

  //
  // upload dialog events
  //
  neko.value!.events.on('upload.dialog.requested', () => {
    dialogRequestActive.value = true
  })
  neko.value!.events.on('upload.dialog.overlay', (id: string) => {
    dialogOverlayActive.value = true
    console.log('upload.dialog.overlay', id)
  })
  neko.value!.events.on('upload.dialog.closed', () => {
    dialogOverlayActive.value = false
    dialogRequestActive.value = false
  })

  //
  // custom messages events
  //
  neko.value!.events.on('receive.unicast', (sender: string, subject: string, body: string) => {
    console.log('receive.unicast', sender, subject, body)
  })
  neko.value!.events.on('receive.broadcast', (sender: string, subject: string, body: string) => {
    console.log('receive.broadcast', sender, subject, body)
  })

  //
  // session events
  //
  neko.value!.events.on('session.created', (id: string) => {
    console.log('session.created', id)
  })
  neko.value!.events.on('session.deleted', (id: string) => {
    console.log('session.deleted', id)
  })
  neko.value!.events.on('session.updated', (id: string) => {
    console.log('session.updated', id)
  })

  //
  // room events
  //
  neko.value!.events.on('room.control.host', (hasHost: boolean, hostID: string | undefined, id: string) => {
    console.log('room.control.host', hasHost, hostID, 'by', id)
  })
  neko.value!.events.on('room.screen.updated', (width: number, height: number, rate: number, id: string) => {
    console.log('room.screen.updated', width, height, rate, 'by', id)
  })
  neko.value!.events.on('room.clipboard.updated', (text: string) => {
    console.log('room.clipboard.updated', text)
  })
  neko.value!.events.on('room.settings.updated', (settings: Settings, id: string) => {
    console.log('room.settings.updated', settings, 'by', id)
  })
  neko.value!.events.on('room.broadcast.status', (isActive: boolean, url?: string) => {
    console.log('room.broadcast.status', isActive, url)
  })

  //
  // control events
  //
  neko.value!.control.on('overlay.click', (e: MouseEvent) => {
    console.log('control: overlay.click', e)
  })
  neko.value!.control.on('overlay.contextmenu', (e: MouseEvent) => {
    console.log('control: overlay.contextmenu', e)
  })

  // custom inactive cursor draw function
  neko.value!.setInactiveCursorDrawFunction(
    (ctx: CanvasRenderingContext2D, x: number, y: number, sessionId: string) => {
      const cursorTag = neko.value!.state.sessions[sessionId]?.profile.name || ''
      const colorLight = '#CCDFF6'
      const colorDark = '#488DDE'

      // get current cursor position
      x -= 4
      y -= 4

      // draw arrow path
      const arrowPath = new Path2D('M5 5L19 12.5L12.3286 14.465L8.29412 20L5 5Z')
      ctx.globalAlpha = 0.5
      ctx.translate(x, y)
      ctx.fillStyle = colorLight
      ctx.fill(arrowPath)
      ctx.lineWidth = 1.5
      ctx.lineJoin = 'miter'
      ctx.miterLimit = 10
      ctx.lineCap = 'round'
      ctx.lineJoin = 'round'
      ctx.strokeStyle = colorDark
      ctx.stroke(arrowPath)

      // draw cursor tag
      if (cursorTag) {
        const x = 20 // box margin x
        const y = 20 // box margin y

        ctx.globalAlpha = 0.5
        ctx.font = '10px Arial, sans-serif'
        ctx.textBaseline = 'top'
        ctx.shadowColor = 'black'
        ctx.shadowBlur = 2
        ctx.lineWidth = 2
        ctx.fillStyle = 'black'
        ctx.strokeText(cursorTag, x, y)
        ctx.shadowBlur = 0
        ctx.fillStyle = 'white'
        ctx.fillText(cursorTag, x, y)
      }
    },
  )

  toggleCursor()
})

const usesCursor = ref(false)

function toggleCursor() {
  if (usesCursor.value) {
    neko.value!.setCursorDrawFunction()
    usesCursor.value = false
    return
  }

  // custom cursor draw function
  neko.value!.setCursorDrawFunction(
    (ctx: CanvasRenderingContext2D, x: number, y: number, {}, {}, sessionId: string) => {
      const cursorTag = neko.value!.state.sessions[sessionId]?.profile.name || ''
      const colorLight = '#CCDFF6'
      const colorDark = '#488DDE'
      const fontColor = '#ffffff'

      // get current cursor position
      x -= 4
      y -= 4

      // draw arrow path
      const arrowPath = new Path2D('M5 5L26 16.5L15.9929 19.513L9.94118 28L5 5Z')
      ctx.translate(x, y)
      ctx.fillStyle = colorLight
      ctx.fill(arrowPath)
      ctx.lineWidth = 2
      ctx.lineJoin = 'miter'
      ctx.miterLimit = 10
      ctx.lineCap = 'round'
      ctx.lineJoin = 'round'
      ctx.strokeStyle = colorDark
      ctx.stroke(arrowPath)

      // draw cursor tag
      if (cursorTag) {
        const fontSize = 12
        const boxPaddingX = 9
        const boxPaddingY = 6

        const x = 22 // box margin x
        const y = 28 // box margin y

        // prepare tag text
        ctx.font = '500 ' + fontSize + 'px Roboto, sans-serif'
        ctx.textBaseline = 'ideographic'

        // create tag container
        const txtWidth = ctx.measureText(cursorTag).width
        const w = txtWidth + boxPaddingX * 2
        const h = fontSize + boxPaddingY * 2
        const r = Math.min(w / 2, h / 2)
        ctx.beginPath()
        ctx.moveTo(x + r, y)
        ctx.arcTo(x + w, y, x + w, y + h, r) // Top-Right
        ctx.arcTo(x + w, y + h, x, y + h, r) // Bottom-Right
        ctx.arcTo(x, y + h, x, y, r * 2) // Bottom-Left
        ctx.arcTo(x, y, x + w, y, r * 2) // Top-Left
        ctx.closePath()
        ctx.fillStyle = colorDark
        ctx.fill()

        // fill in tag text
        ctx.fillStyle = fontColor
        ctx.fillText(cursorTag, x + boxPaddingX, y + fontSize + boxPaddingY)
      }
    },
  )

  usesCursor.value = true
}
</script>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Roboto:wght@400;500&display=swap');
</style>
