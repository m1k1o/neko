<template>
  <div>
    <div style="float: right; max-width: 500px">
      <h3>State</h3>
      <table class="states" v-if="loaded">
        <tr class="ok">
          <th>connection.websocket</th>
          <td>{{ neko.state.connection.websocket }}</td>
        </tr>
        <tr class="ok">
          <th>connection.webrtc</th>
          <td>{{ neko.state.connection.webrtc }}</td>
        </tr>
        <tr>
          <th>connection.type</th>
          <td>{{ neko.state.connection.type }}</td>
        </tr>
        <tr>
          <th>connection.can_watch</th>
          <td>{{ neko.state.connection.can_watch }}</td>
        </tr>
        <tr>
          <th>connection.can_control</th>
          <td>{{ neko.state.connection.can_control }}</td>
        </tr>
        <tr>
          <th>connection.clipboard_access</th>
          <td>{{ neko.state.connection.clipboard_access }}</td>
        </tr>
        <tr class="ok">
          <th>video.playable</th>
          <td>{{ neko.state.video.playable }}</td>
        </tr>
        <tr class="ok">
          <th>video.playing</th>
          <td>{{ neko.state.video.playing }}</td>
        </tr>
        <tr>
          <th>video.volume</th>
          <td>{{ neko.state.video.volume }}</td>
        </tr>
        <tr class="ok">
          <th>control.scroll.inverse</th>
          <td><input type="checkbox" v-model="neko.state.control.scroll.inverse" /></td>
        </tr>
        <tr class="ok">
          <th>control.scroll.sensitivity</th>
          <td><input type="number" v-model="neko.state.control.scroll.sensitivity" /></td>
        </tr>
        <tr>
          <th>control.host</th>
          <td>{{ neko.state.control.host }}</td>
        </tr>
        <tr class="ok">
          <th>screen.size.width</th>
          <td>{{ neko.state.screen.size.width }}</td>
        </tr>
        <tr class="ok">
          <th>screen.size.height</th>
          <td>{{ neko.state.screen.size.height }}</td>
        </tr>
        <tr class="ok">
          <th>screen.size.rate</th>
          <td>{{ neko.state.screen.size.rate }}</td>
        </tr>
        <tr class="ok">
          <th>screen.configurations</th>
          <td>
            <select
              :value="Object.values(neko.state.screen.size).join()"
              @input="
                a = String($event.target.value).split(',')
                neko.screen.size(parseInt(a[0]), parseInt(a[1]), parseInt(a[2]))
              "
            >
              <option
                v-for="{ width, height, rate } in neko.state.screen.configurations"
                :key="width + height + rate"
                :value="[width, height, rate].join()"
              >
                {{ width }}x{{ height }}@{{ rate }}
              </option>
            </select>
          </td>
        </tr>
        <tr>
          <th>screen.is_fullscreen</th>
          <td>{{ neko.state.screen.is_fullscreen }}</td>
        </tr>
        <tr class="ok">
          <th>member.id</th>
          <td>{{ neko.state.member.id }}</td>
        </tr>
        <tr class="ok">
          <th>member.name</th>
          <td>{{ neko.state.member.name }}</td>
        </tr>
        <tr>
          <th>member.is_admin</th>
          <td>{{ neko.state.member.is_admin }}</td>
        </tr>
        <tr>
          <th>member.is_watching</th>
          <td>{{ neko.state.member.is_watching }}</td>
        </tr>
        <tr class="ok">
          <th>member.is_controlling</th>
          <td>{{ neko.state.member.is_controlling }}</td>
        </tr>
        <tr>
          <th>member.can_watch</th>
          <td>{{ neko.state.member.can_watch }}</td>
        </tr>
        <tr>
          <th>member.can_control</th>
          <td>{{ neko.state.member.can_control }}</td>
        </tr>
        <tr>
          <th>member.clipboard_access</th>
          <td>{{ neko.state.member.clipboard_access }}</td>
        </tr>
        <tr>
          <th>members</th>
          <td>{{ neko.state.members }}</td>
        </tr>
      </table>
    </div>
    <div>
      <button @click="connect()">Connect</button>
      <button @click="disconnect()">Disonnect</button>

      <template v-if="loaded && neko.connected">
        <button v-if="!is_controlling" @click="neko.control.request()">request control</button>
        <button v-else @click="neko.control.release()">release control</button>

        <button @click="neko.video.pause()">pause stream</button>
        <button @click="neko.video.play()">play stream</button><br />
      </template>

      <div ref="container" style="width: 1280px; height: 720px; border: 2px solid red">
        <neko-canvas ref="neko" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .states {
    td,
    th {
      border: 1px solid black;
      padding: 4px;
    }

    th {
      text-align: left;
    }

    .ok {
      background: #97f197;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'

  import Neko from '~/components/canvas.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-canvas': Neko,
    },
  })
  export default class extends Vue {
    @Ref('container') readonly container!: HTMLElement
    @Ref('neko') readonly neko!: Neko
    loaded: boolean = false

    get is_controlling() {
      return this.neko.state.member.is_controlling
    }

    connect() {
      this.neko.connect('ws://192.168.1.20:3000/', 'admin', 'test')
    }

    disconnect() {
      this.neko.disconnect()
    }

    mounted() {
      this.loaded = true

      this.neko.events.on('system.websocket', (status) => {
        console.log('system.websocket', status)
      })
      this.neko.events.on('system.webrtc', (status) => {
        console.log('system.webrtc', status)
      })
      this.neko.events.on('system.connect', () => {
        console.log('system.connect')
      })
      this.neko.events.on('system.disconnect', (message) => {
        console.log('system.disconnect', message)
      })
      this.neko.events.on('member.list', (members) => {
        console.log('member.list', members)
      })
      this.neko.events.on('member.connected', (id) => {
        console.log('member.connected', id)
      })
      this.neko.events.on('member.disconnected', (id) => {
        console.log('member.disconnected', id)
      })
      this.neko.events.on('control.host', (id) => {
        console.log('control.host', id)
      })
      this.neko.events.on('control.request', (id) => {
        console.log('control.request', id)
      })
      this.neko.events.on('control.requesting', (id) => {
        console.log('control.requesting', id)
      })
      this.neko.events.on('clipboard.update', (text) => {
        console.log('clipboard.update', text)
      })
      this.neko.events.on('screen.size', (id) => {
        console.log('screen.size', id)
      })
      this.neko.events.on('broadcast.status', (payload) => {
        console.log('broadcast.status', payload)
      })
      this.neko.events.on('member.ban', (id, target) => {
        console.log('member.ban', id, target)
      })
      this.neko.events.on('member.kick', (id, target) => {
        console.log('member.kick', id, target)
      })
      this.neko.events.on('member.muted', (id, target) => {
        console.log('member.muted', id, target)
      })
      this.neko.events.on('member.unmuted', (id, target) => {
        console.log('member.unmuted', id, target)
      })
      this.neko.events.on('room.locked', (id) => {
        console.log('room.locked', id)
      })
      this.neko.events.on('room.unlocked', (id) => {
        console.log('room.unlocked', id)
      })
    }
  }
</script>
