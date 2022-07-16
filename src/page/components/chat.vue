<template>
  <div class="chat">
    <ul class="chat-history" ref="history">
      <template v-for="(message, index) in history">
        <li :key="index" class="message" v-show="neko && neko.connected">
          <div class="content">
            <div class="content-head">
              <span class="session">{{ session(message.id) }}</span>
              <span class="timestamp">{{ timestamp(message.created) }}</span>
            </div>
            <p>{{ message.content }}</p>
          </div>
        </li>
      </template>
    </ul>
    <div class="chat-send">
      <div class="text-container">
        <textarea ref="input" placeholder="Send a message" @keydown="onKeyDown" v-model="content" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  @import '@/page/assets/styles/main.scss';

  .chat {
    flex: 1;
    flex-direction: column;
    display: flex;
    max-height: 100%;
    max-width: 100%;
    overflow-x: hidden;

    .chat-history {
      flex: 1;
      overflow-y: scroll;
      overflow-x: hidden;
      max-width: 100%;
      scrollbar-width: thin;
      scrollbar-color: $background-tertiary transparent;

      &::-webkit-scrollbar {
        width: 8px;
      }

      &::-webkit-scrollbar-track {
        background-color: transparent;
      }

      &::-webkit-scrollbar-thumb {
        background-color: $background-tertiary;
        border: 2px solid $background-primary;
        border-radius: 4px;
      }

      &::-webkit-scrollbar-thumb:hover {
        background-color: $background-floating;
      }

      .message {
        flex: 1;
        border-top: 1px solid var(--border-color);
        padding: 10px 5px 0px 10px;
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
        overflow: hidden;
        user-select: text;
        word-wrap: break-word;
        font-size: 16px;
      }

      .content-head {
        cursor: default;
        width: 100%;
        margin-bottom: 3px;
        display: block;

        .session {
          display: inline-block;
          color: $text-normal;
          font-weight: bold;
        }

        .timestamp {
          color: $text-muted;
          font-size: 0.7rem;
          margin-left: 0.3rem;
          line-height: 12px;
          &::first-letter {
            text-transform: uppercase;
          }
        }
      }
    }

    .chat-send {
      flex-shrink: 0;
      height: 80px;
      max-height: 80px;
      padding: 0 10px 10px 10px;
      flex-direction: column;
      display: flex;

      .text-container {
        flex: 1;
        width: 100%;
        height: 100%;
        background-color: rgba($color: #fff, $alpha: 0.05);
        border-radius: 5px;
        position: relative;
        display: flex;

        .emoji-menu {
          width: 20px;
          height: 20px;
          font-size: 20px;
          margin: 8px 5px 0 0;
          cursor: pointer;
        }

        textarea {
          flex: 1;
          font-family: $text-family;
          border: none;
          caret-color: $text-normal;
          color: $text-normal;
          resize: none;
          margin: 5px;
          background-color: transparent;
          scrollbar-width: thin;
          scrollbar-color: $background-tertiary transparent;

          &::placeholder {
            color: $text-muted;
          }

          &::-webkit-scrollbar {
            width: 4px;
          }

          &::-webkit-scrollbar-track {
            background-color: transparent;
          }

          &::-webkit-scrollbar-thumb {
            background-color: $background-tertiary;
            border-radius: 4px;
          }

          &::-webkit-scrollbar-thumb:hover {
            background-color: $background-floating;
          }

          &::selection {
            background: $text-link;
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop, Watch, Ref } from 'vue-property-decorator'
  import Neko from '~/component/main.vue'

  const length = 512 // max length of message

  @Component({
    name: 'neko-chat',
  })
  export default class extends Vue {
    @Ref('history') readonly _history!: HTMLElement
    @Prop() readonly neko!: Neko

    history = []
    content = ''

    mounted() {
      this.$nextTick(() => {
        this._history.scrollTop = this._history.scrollHeight
      })
    }

    timestamp(date: Date | string) {
      date = new Date(date)

      return (
        date.getFullYear() +
        '-' +
        String(date.getMonth() + 1).padStart(2, '0') +
        '-' +
        String(date.getDate()).padStart(2, '0') +
        ' ' +
        String(date.getHours()).padStart(2, '0') +
        ':' +
        String(date.getMinutes()).padStart(2, '0') +
        ':' +
        String(date.getSeconds()).padStart(2, '0')
      )
    }

    session(id: string) {
      let session = this.neko.state.sessions[id]
      return session ? session.profile.name : id
    }

    @Watch('neko')
    onNekoChange() {
      this.neko.events.on('receive.broadcast', (sender: string, subject: string, body: string) => {
        if (subject === 'chat') {
          Vue.set(this, 'history', [...this.history, body])
        }
      })
    }

    @Watch('history')
    onHistroyChange() {
      this.$nextTick(() => {
        this._history.scrollTop = this._history.scrollHeight
      })
    }

    onKeyDown(event: KeyboardEvent) {
      if (this.content.length > length) {
        this.content = this.content.substring(0, length)
      }

      if (this.content.length == length) {
        if (
          [8, 16, 17, 18, 20, 33, 34, 35, 36, 37, 38, 39, 40, 45, 46, 91, 93, 144].includes(event.keyCode) ||
          (event.ctrlKey && [67, 65, 88].includes(event.keyCode))
        ) {
          return
        }
        event.preventDefault()
        return
      }

      if (event.keyCode !== 13 || event.shiftKey) {
        return
      }

      if (this.content === '') {
        event.preventDefault()
        return
      }

      this.$emit('send_message', this.content)

      let message = {
        id: this.neko.state.session_id,
        created: new Date(),
        content: this.content,
      }

      this.neko.sendBroadcast('chat', message)
      Vue.set(this, 'history', [...this.history, message])

      this.content = ''
      event.preventDefault()
    }
  }
</script>
