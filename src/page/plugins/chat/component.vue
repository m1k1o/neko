<template>
  <div class="chat" v-if="tab === 'chat'">
    <div class="chat-header">
      <p>Chat sending is {{ canSendInSettings ? 'unlocked' : 'locked' }} for users</p>
      <i v-if="props.neko.is_admin" :class="['fas', canSendInSettings ? 'fa-unlock' : 'fa-lock', 'refresh']" @click="toggleCanSend(!canSendInSettings)" :title="canSendInSettings ? 'Lock' : 'Unlock'" />
    </div>
    <div class="chat-header">
      <p>Chat receiving is {{ canReceiveInSettings ? 'unlocked' : 'locked' }} for users</p>
      <i v-if="props.neko.is_admin" :class="['fas', canReceiveInSettings ? 'fa-unlock' : 'fa-lock', 'refresh']" @click="toggleCanReceive(!canReceiveInSettings)" :title="canReceiveInSettings ? 'Lock' : 'Unlock'" />
    </div>
    <div v-if="!enabledSystemWide" class="chat-header">
      <p>Chat is disabled system-wide</p>
    </div>
    <ul class="chat-history" ref="history">
      <template v-for="(message, index) in messages" :key="index">
        <li class="message" v-show="neko && neko.connected">
          <div class="content">
            <div class="content-head">
              <span class="session">{{ session(message.id) }}</span>
              <span class="timestamp">{{ timestamp(message.created) }}</span>
            </div>
            <p>{{ message.content.text }}</p>
          </div>
        </li>
      </template>
    </ul>
    <div class="chat-send" v-if="canSend">
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

    .chat-header {
      display: flex;
      flex-direction: row;
      margin: 10px 10px 0px 10px;
      padding: 0.5em;
      font-weight: 600;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;

      .refresh {
        margin-left: auto;
        cursor: pointer;
      }
    }

    .chat-history {
      flex: 1;
      overflow-y: scroll;
      overflow-x: hidden;
      max-width: 100%;
      scrollbar-width: thin;
      scrollbar-color: $background-tertiary transparent;
      margin: 10px 0;
      box-sizing: border-box;

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
        line-height: 1.2;
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
      padding: 10px 10px 10px 10px;
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

<script lang="ts" setup>
import { ref, watch, computed, onMounted } from 'vue'
import type Neko from '@/component/main.vue'
import * as types from './types'

// TODO: Use API.
// import { ChatApi } from './api'
// const api = props.neko.withApi(ChatApi) as ChatApi

const length = 512 // max length of message
const history = ref<HTMLUListElement | null>(null)

const props = defineProps<{
  neko: typeof Neko,
  tab: string
}>()

// config option to enable/disable chat plugin
const enabledSystemWide = ref(false)
// dynamic settings for chat plugin
const canSendInSettings = computed(() => !(props.neko.state.settings?.plugins?.['chat.can_send'] === false))
const canReceiveInSettings = computed(() => !(props.neko.state.settings?.plugins?.['chat.can_receive'] === false))
// user specific setting to enable/disable chat plugin
const canSendForMe = computed(() => !(props.neko.session?.profile?.plugins?.['chat.can_send'] === false))
const canReceiveForMe = computed(() => !(props.neko.session?.profile?.plugins?.['chat.can_receive'] === false))
// combined enabled state for chat plugin and user
const canSend = computed(() => enabledSystemWide.value && (canSendInSettings.value || props.neko.is_admin) && canSendForMe.value)
const canReceive = computed(() => enabledSystemWide.value && (canReceiveInSettings.value || props.neko.is_admin) && canReceiveForMe.value)

const messages = ref<types.Message[]>([])
const content = ref('')

onMounted(() => {
  props.neko.events.on('message', async (event: string, payload: any) => {
    switch (event) {
      case types.CHAT_INIT: {
        const message = payload as types.Init
        enabledSystemWide.value = message.enabled
        break
      }
      case types.CHAT_MESSAGE: {
        const message = payload as types.Message
        messages.value = [...messages.value, message]
        break
      }
    }
  })

  setTimeout(() => {
    if (history.value)
      history.value.scrollTop = history.value.scrollHeight
  }, 0)
})

async function toggleCanSend(isEnabled = true) {
  try {
    await props.neko.room.settingsSet({ plugins: { "chat.can_send": isEnabled } })
  } catch (e: any) {
    alert(e.response ? e.response.data.message : e)
  }
}

async function toggleCanReceive(isEnabled = true) {
  try {
    await props.neko.room.settingsSet({ plugins: { "chat.can_receive": isEnabled } })
  } catch (e: any) {
    alert(e.response ? e.response.data.message : e)
  }
}

function timestamp(date: Date | string) {
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

function session(id: string) {
  let session = props.neko.state.sessions[id]
  return session ? session.profile.name : id
}

watch(messages, function() {
  setTimeout(() => {
    if (history.value)
      history.value.scrollTop = history.value.scrollHeight
  }, 0)
})

function onKeyDown(event: KeyboardEvent) {
  if (content.value.length > length) {
    content.value = content.value.substring(0, length)
  }

  if (content.value.length == length) {
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
  
  sendMessage()
  event.preventDefault()
}

function sendMessage() {
  if (content.value === '') {
    return
  }

  props.neko.sendMessage(types.CHAT_MESSAGE, {
    text: content.value,
  } as types.Content)

  content.value = ''
}
</script>
