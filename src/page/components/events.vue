<template>
  <div class="tab-states">
    <table class="states">
      <tr>
        <th style="width: 50%">authenticated</th>
        <td :style="!neko.state.authenticated ? 'background: red;' : ''">{{ neko.state.authenticated }}</td>
      </tr>
      <tr>
        <th>connection.url</th>
        <td style="word-break: break-all">{{ neko.state.connection.url }}</td>
      </tr>
      <tr>
        <th>connection.token</th>
        <td>{{ neko.state.connection.token ? 'yes' : 'no' }}</td>
      </tr>
      <tr>
        <th>connection.status</th>
        <td
          :style="
            neko.state.connection.status == 'disconnected'
              ? 'background: red;'
              : neko.state.connection.status == 'connecting'
                ? 'background: #17448a;'
                : ''
          "
        >
          {{ neko.state.connection.status }}
        </td>
      </tr>
      <tr>
        <th title="connection.websocket.connected">connection.websocket.con...</th>
        <td>{{ neko.state.connection.websocket.connected }}</td>
      </tr>
      <tr>
        <th>connection.websocket.config</th>
        <td>
          <details>
            <summary>Show</summary>
            <table class="states">
              <tr>
                <th style="width: 40%">max_reconnects</th>
                <td>
                  {{ neko.state.connection.websocket.config.max_reconnects }}
                </td>
              </tr>
              <tr>
                <th style="width: 40%">timeout_ms</th>
                <td>
                  {{ neko.state.connection.websocket.config.timeout_ms }}
                </td>
              </tr>
              <tr>
                <th style="width: 40%">backoff_ms</th>
                <td>
                  {{ neko.state.connection.websocket.config.backoff_ms }}
                </td>
              </tr>
            </table>
          </details>
        </td>
      </tr>
      <tr>
        <th title="connection.webrtc.connected">connection.webrtc.connect...</th>
        <td>{{ neko.state.connection.webrtc.connected }}</td>
      </tr>
      <tr>
        <th>connection.webrtc.stable</th>
        <td>{{ neko.state.connection.webrtc.stable }}</td>
      </tr>
      <tr>
        <th>connection.webrtc.config</th>
        <td>
          <details>
            <summary>Show</summary>
            <table class="states">
              <tr>
                <th style="width: 40%">max_reconnects</th>
                <td>
                  {{ neko.state.connection.webrtc.config.max_reconnects }}
                </td>
              </tr>
              <tr>
                <th style="width: 40%">timeout_ms</th>
                <td>
                  {{ neko.state.connection.webrtc.config.timeout_ms }}
                </td>
              </tr>
              <tr>
                <th style="width: 40%">backoff_ms</th>
                <td>
                  {{ neko.state.connection.webrtc.config.backoff_ms }}
                </td>
              </tr>
            </table>
          </details>
        </td>
      </tr>
      <tr>
        <th>connection.webrtc.stats</th>
        <td>
          <table class="states" v-if="neko.state.connection.webrtc.stats != null">
            <tr>
              <th style="width: 40%">muted</th>
              <td :style="neko.state.connection.webrtc.stats.muted ? 'background: red' : ''">
                {{ neko.state.connection.webrtc.stats.muted }}
              </td>
            </tr>
            <tr>
              <th style="width: 40%">bitrate</th>
              <td>{{ Math.floor(neko.state.connection.webrtc.stats.bitrate / 1024 / 8) }} KB/s</td>
            </tr>
            <tr>
              <th style="width: 40%">latency</th>
              <td
                :title="
                  'request: ' +
                  neko.state.connection.webrtc.stats.requestLatency +
                  'ms, response: ' +
                  neko.state.connection.webrtc.stats.responseLatency +
                  'ms'
                "
              >
                {{ neko.state.connection.webrtc.stats.latency }}ms
              </td>
            </tr>
            <tr>
              <th>loss</th>
              <td :style="neko.state.connection.webrtc.stats.packetLoss >= 1 ? 'background: red' : ''">
                {{ Math.floor(neko.state.connection.webrtc.stats.packetLoss) }}%
              </td>
            </tr>
            <tr>
              <td
                colspan="2"
                style="background: green; text-align: center"
                v-if="neko.state.connection.webrtc.stats.paused"
              >
                webrtc is paused
              </td>
              <td
                colspan="2"
                style="background: darkviolet; text-align: center"
                v-else-if="neko.state.connection.webrtc.video.disabled"
              >
                video is disabled
              </td>
              <td
                colspan="2"
                style="background: red; text-align: center"
                v-else-if="!neko.state.connection.webrtc.stats.fps"
              >
                frame rate is zero
              </td>
              <td colspan="2" v-else>
                {{
                  neko.state.connection.webrtc.stats.width +
                  'x' +
                  neko.state.connection.webrtc.stats.height +
                  '@' +
                  Math.floor(neko.state.connection.webrtc.stats.fps * 100) / 100
                }}
              </td>
            </tr>
          </table>
        </td>
      </tr>
      <tr>
        <th title="connection.webrtc.video.disabled">connection.webrtc.video.disab..</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.connection.webrtc.video.disabled }}</span>
            <button @click="neko.setWebRTCVideo({ disabled: !neko.state.connection.webrtc.video.disabled })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th>connection.webrtc.video.id</th>
        <td>{{ neko.state.connection.webrtc.video.id }}</td>
      </tr>
      <tr>
        <th>connection.webrtc.video.auto</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.connection.webrtc.video.auto }}</span>
            <button @click="neko.setWebRTCVideo({ auto: !neko.state.connection.webrtc.video.auto })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th title="connection.webrtc.audio.disabled">connection.webrtc.audio.disab..</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.connection.webrtc.audio.disabled }}</span>
            <button @click="neko.setWebRTCAudio({ disabled: !neko.state.connection.webrtc.audio.disabled })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th rowspan="2">connection.webrtc.videos</th>
        <td>Total {{ neko.state.connection.webrtc.videos.length }} videos.</td>
      </tr>
      <tr>
        <td>
          <select
            :value="neko.state.connection.webrtc.video.id"
            @input="neko.setWebRTCVideo({ selector: { id: ($event.target as HTMLSelectElement)!.value || '' } })"
          >
            <option v-for="video in neko.state.connection.webrtc.videos" :key="video" :value="video">
              {{ video }}
            </option>
          </select>
        </td>
      </tr>
      <tr>
        <th>connection.screencast</th>
        <td>{{ neko.state.connection.screencast }}</td>
      </tr>
      <tr>
        <th>connection.type</th>
        <td :style="neko.state.connection.type == 'fallback' ? 'background: #17448a;' : ''">
          {{ neko.state.connection.type }}
        </td>
      </tr>
      <tr>
        <th>video.playable</th>
        <td>{{ neko.state.video.playable }}</td>
      </tr>
      <tr>
        <th rowspan="2">video.playing</th>
        <td>{{ neko.state.video.playing }}</td>
      </tr>
      <tr>
        <td>
          <button v-if="!neko.state.video.playing" @click="neko.play()">play</button>
          <button v-else @click="neko.pause()">pause</button>
        </td>
      </tr>
      <tr>
        <th rowspan="2">video.volume</th>
        <td>{{ neko.state.video.volume }}</td>
      </tr>
      <tr>
        <td>
          <input
            type="range"
            min="0"
            max="1"
            :value="neko.state.video.volume"
            @input="neko.setVolume(Number(($event.target as HTMLInputElement)!.value))"
            step="0.01"
          />
        </td>
      </tr>
      <tr>
        <th rowspan="2">video.muted</th>
        <td>{{ neko.state.video.muted }}</td>
      </tr>
      <tr>
        <td>
          <button v-if="!neko.state.video.muted" @click="neko.mute()">mute</button>
          <button v-else @click="neko.unmute()">unmute</button>
        </td>
      </tr>
      <tr>
        <th>control.scroll.inverse</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.control.scroll.inverse }}</span>
            <button @click="neko.setScrollInverse(!neko.state.control.scroll.inverse)">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th rowspan="2">control.scroll.sensitivity</th>
        <td>{{ neko.state.control.scroll.sensitivity }}</td>
      </tr>
      <tr>
        <td>
          <input
            type="range"
            min="-5"
            max="5"
            :value="neko.state.control.scroll.sensitivity"
            @input="neko.setScrollSensitivity(Number(($event.target as HTMLInputElement)!.value))"
            step="1"
          />
        </td>
      </tr>
      <tr>
        <th>control.clipboard</th>
        <td>
          <textarea
            :readonly="!neko.controlling"
            :value="neko.state.control.clipboard ? neko.state.control.clipboard.text : ''"
            @input="clipboardText = ($event.target as HTMLTextAreaElement)!.value"
          ></textarea>
          <button :disabled="!neko.controlling" @click="neko.room.clipboardSetText({ text: clipboardText })">
            send clipboard
          </button>
        </td>
      </tr>
      <tr>
        <th rowspan="2">control.keyboard</th>
        <td>
          {{
            neko.state.control.keyboard.layout +
            (neko.state.control.keyboard.variant ? ' (' + neko.state.control.keyboard.variant + ')' : '')
          }}
        </td>
      </tr>
      <tr>
        <td>
          <input
            type="text"
            placeholder="Layout"
            :value="neko.state.control.keyboard.layout"
            @input="neko.setKeyboard(($event.target as HTMLInputElement)!.value, neko.state.control.keyboard.variant)"
            style="width: 50%; box-sizing: border-box"
          />
          <input
            type="text"
            placeholder="Variant"
            :value="neko.state.control.keyboard.variant"
            @input="neko.setKeyboard(neko.state.control.keyboard.layout, ($event.target as HTMLInputElement)!.value)"
            style="width: 50%; box-sizing: border-box"
          />
        </td>
      </tr>
      <tr>
        <th>control.touch.enabled</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.control.touch.enabled }}</span>
            <button @click="neko.setTouchEnabled(!neko.state.control.touch.enabled)">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th>control.touch.supported</th>
        <td>{{ neko.state.control.touch.supported }}</td>
      </tr>
      <tr>
        <th rowspan="2">control.host_id</th>
        <td>{{ neko.state.control.host_id }}</td>
      </tr>
      <tr>
        <td>
          <button v-if="!neko.controlling" @click="neko.room.controlRequest()">request control</button>
          <button v-else @click="neko.room.controlRelease()">release control</button>
        </td>
      </tr>
      <tr>
        <th>screen.size</th>
        <td>
          {{ neko.state.screen.size.width }}x{{ neko.state.screen.size.height }}@{{ neko.state.screen.size.rate }}
        </td>
      </tr>
      <template v-if="neko.is_admin">
        <tr>
          <th rowspan="2">screen.configurations</th>
          <td>Total {{ neko.state.screen.configurations.length }} configurations.</td>
        </tr>
        <tr>
          <td>
            <input
              list="screen-configuration"
              v-model="screenConfiguration"
              style="width: 100%; box-sizing: border-box"
            />
            <datalist id="screen-configuration">
              <option
                v-for="{ width, height, rate } in neko.state.screen.configurations"
                :key="String(width) + 'x' + String(height) + '@' + String(rate)"
                :value="String(width) + 'x' + String(height) + '@' + String(rate)"
              />
            </datalist>
            <button @click="setScreenConfiguration">set</button>
          </td>
        </tr>
        <tr>
          <th class="middle">screen.sync.enabled</th>
          <td>
            <div class="space-between">
              <span>{{ neko.state.screen.sync.enabled }}</span>
              <button @click="neko.state.screen.sync.enabled = !neko.state.screen.sync.enabled">
                <i class="fas fa-toggle-on"></i>
              </button>
            </div>
          </td>
        </tr>
        <tr>
          <th rowspan="2">screen.sync.multiplier</th>
          <td>{{ neko.state.screen.sync.multiplier || 'use device pixel ratio' }}</td>
        </tr>
        <tr>
          <td>
            <input
              type="range"
              min="0"
              max="10"
              :value="neko.state.screen.sync.multiplier"
              @input="neko.state.screen.sync.multiplier = Number(($event.target as HTMLInputElement)!.value)"
              step="0.1"
            />
          </td>
        </tr>
        <tr>
          <th rowspan="2">screen.sync.rate</th>
          <td>{{ neko.state.screen.sync.rate }}</td>
        </tr>
        <tr>
          <td>
            <input
              type="range"
              min="5"
              max="60"
              :value="neko.state.screen.sync.rate"
              @input="neko.state.screen.sync.rate = Number(($event.target as HTMLInputElement)!.value)"
              step="5"
            />
          </td>
        </tr>
      </template>
      <template v-else>
        <tr>
          <th>screen.configurations</th>
          <td rowspan="2" style="vertical-align: middle">Session is not admin.</td>
        </tr>
        <tr>
          <th>screen.sync</th>
        </tr>
      </template>
      <tr>
        <th>session_id</th>
        <td>{{ neko.state.session_id }}</td>
      </tr>
      <tr>
        <th>sessions</th>
        <td>Total {{ Object.values(neko.state.sessions).length }} sessions.</td>
      </tr>

      <tr>
        <th class="middle">settings.private_mode</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.private_mode }}</span>
            <button @click="updateSettings({ private_mode: !neko.state.settings.private_mode })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th class="middle">settings.locked_logins</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.locked_logins }}</span>
            <button @click="updateSettings({ locked_logins: !neko.state.settings.locked_logins })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th class="middle">settings.locked_controls</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.locked_controls }}</span>
            <button @click="updateSettings({ locked_controls: !neko.state.settings.locked_controls })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th class="middle">settings.control_protection</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.control_protection }}</span>
            <button @click="updateSettings({ control_protection: !neko.state.settings.control_protection })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th class="middle">settings.implicit_hosting</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.implicit_hosting }}</span>
            <button @click="updateSettings({ implicit_hosting: !neko.state.settings.implicit_hosting })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th class="middle">settings.inactive_cursors</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.inactive_cursors }}</span>
            <button @click="updateSettings({ inactive_cursors: !neko.state.settings.inactive_cursors })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th class="middle">settings.merciful_reconnect</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.settings.merciful_reconnect }}</span>
            <button @click="updateSettings({ merciful_reconnect: !neko.state.settings.merciful_reconnect })">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>

      <tr>
        <th>cursors</th>
        <td>{{ neko.state.cursors }}</td>
      </tr>

      <tr>
        <th>mobile_keyboard_open</th>
        <td>
          <div class="space-between">
            <span>{{ neko.state.mobile_keyboard_open }}</span>
            <button @click="neko.mobileKeyboardToggle">
              <i class="fas fa-toggle-on"></i>
            </button>
          </div>
        </td>
      </tr>

      <tr>
        <th>control actions</th>
        <td>
          <button title="cut" @click="neko.control.cut()"><i class="fas fa-cut" /></button>
          <button title="copy" @click="neko.control.copy()"><i class="fas fa-copy" /></button>
          <button title="paste" @click="neko.control.paste()"><i class="fas fa-paste" /></button>
          <button title="select all" @click="neko.control.selectAll()"><i class="fas fa-i-cursor" /></button>
        </td>
      </tr>
      <tr>
        <th>control keypress</th>
        <td style="text-align: center">
          <button style="width: 20px" v-for="l in letters" :key="l" @click="neko.control.keyPress(l)">
            {{ String.fromCharCode(l) }}
          </button>
          <div style="display: flex">
            <button title="shift" @click="shift = !shift">
              <i v-if="shift" class="fas fa-caret-square-up" />
              <i v-else class="far fa-caret-square-up" />
            </button>
            <button style="width: 100%" @click="neko.control.keyPress(' '.charCodeAt(0))">space</button>
            <button title="shift" @click="shift = !shift">
              <i v-if="shift" class="fas fa-caret-square-up" />
              <i v-else class="far fa-caret-square-up" />
            </button>
          </div>
        </td>
      </tr>
      <tr>
        <th>chaos monkey</th>
        <td>
          <button @click="cursorMovingToggle">cursor moving</button>
          <button @click="screenChangingToggle">screen cfg changing</button>
        </td>
      </tr>
      <tr>
        <th>click on color</th>
        <td>
          <NekoColor @colorChange="neko.control.buttonPress(1, $event)" />
        </td>
      </tr>
    </table>
  </div>
</template>

<style lang="scss">
  .tab-states {
    &,
    .states {
      width: 100%;
    }

    td,
    th {
      border: 1px solid #ccc;
      padding: 4px;
    }

    th {
      text-align: left;
    }

    .middle {
      vertical-align: middle;
    }

    .space-between {
      width: 100%;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }
</style>

<script lang="ts" setup>
import { ref, watch, computed } from 'vue'
import type Neko from '@/component/main.vue'
import NekoColor from './color.vue'

const props = defineProps<{
  neko: typeof Neko
}>()

const clipboardText = ref('')
const bitrate = ref<number | null>(null)

watch(() => props.neko.state.connection.webrtc.bitrate, (val) => {
  bitrate.value = val
})

const shift = ref(false)
const letters = computed(() => {
  let letters = [] as number[]
  for (let i = (shift.value ? 'A' : 'a').charCodeAt(0); i <= (shift.value ? 'Z' : 'z').charCodeAt(0); i++) {
    letters.push(i)
  }
  return letters
})

// fast sceen changing test
let screen_interval: number | null = null
function screenChangingToggle() {
  if (screen_interval === null) {
    let sizes = props.neko.state.screen.configurations
    let len = sizes.length

    screen_interval = window.setInterval(() => {
      let { width, height, rate } = sizes[Math.floor(Math.random() * len)]

      props.neko.setScreenSize(width, height, rate)
    }, 10)
  } else {
    window.clearInterval(screen_interval)
    screen_interval = null
  }
}

const screenConfiguration = ref('')
function setScreenConfiguration() {
  let [width, height, rate] = screenConfiguration.value.split(/[@x]/)
  props.neko.setScreenSize(parseInt(width), parseInt(height), parseInt(rate))
}

watch(() => props.neko.state.screen.size, (val) => {
  screenConfiguration.value = `${val.width}x${val.height}@${val.rate}`
})

// fast cursor moving test
let cursor_interval: number | null = null
function cursorMovingToggle() {
  if (cursor_interval === null) {
    let len = props.neko.state.screen.size.width

    cursor_interval = window.setInterval(() => {
      let x = Math.floor(Math.random() * len)
      let y = Math.floor(Math.random() * len)

      props.neko.control.move({ x, y })
    }, 10)
  } else {
    window.clearInterval(cursor_interval)
    cursor_interval = null
  }
}

async function updateSettings(settings: any) {
  try {
    await props.neko.room.settingsSet(settings)
  } catch (e: any) {
    alert(e.response ? e.response.data.message : e)
  }
}
</script>
