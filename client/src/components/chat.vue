<template>
  <div class="chat">
    <ul class="chat-history" ref="history" @click="onClick">
      <template v-for="(message, index) in history">
        <li :key="index" class="message" v-if="message.type === 'text'">
          <div class="author" @contextmenu.stop.prevent="onContext($event, { member: member(message.id) })">
            <img :src="`https://api.adorable.io/avatars/40/${member(message.id).username}.png`" />
          </div>
          <div class="content">
            <div class="content-head">
              <span>{{ member(message.id).username }}</span>
              <span class="timestamp">{{ timestamp(message.created) }}</span>
            </div>
            <div class="content-body">
              <neko-markdown :source="message.content" />
            </div>
          </div>
        </li>
        <li :key="index" class="event" v-if="message.type === 'event'">
          <span
            v-tooltip="{
              content: `${timestamp(message.created)}, ${member(message.id).username} ${message.content}`,
              placement: 'left',
              offset: 3,
              boundariesElement: 'body',
            }"
          >
            <strong v-if="message.id === id">You</strong>
            <strong v-else>{{ member(message.id).username }}</strong>
            {{ message.content }}
          </span>
        </li>
      </template>
    </ul>
    <neko-context ref="context" />
    <div v-if="!muted" class="chat-send">
      <div class="accent" />
      <div class="text-container">
        <textarea ref="chat" placeholder="Send a message" @keydown="onKeyDown" v-model="content" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
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

      ::v-deep *::selection {
        background: $text-link;
      }

      li {
        flex: 1;
        border-top: 1px solid var(--border-color);
        padding: 10px 10px 0px 10px;
        display: flex;
        flex-direction: row;
        overflow: hidden;
        user-select: contain;
        max-width: 100%;

        &.message {
          .author {
            flex-grow: 0;
            flex-shrink: 0;
            overflow: hidden;
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: $style-primary;
            margin: 5px 10px 10px 0px;

            img {
              width: 100%;
            }
          }

          .content {
            flex: 1;
            display: flex;
            flex-direction: column;
            line-height: 22px;

            .content-head {
              cursor: default;

              span {
                color: $text-normal;
                font-weight: 500;
                float: left;
              }

              .timestamp {
                margin-left: 5px;
                color: $text-muted;
                float: left;
                font-size: 0.8em;
                line-height: 1.7em;
              }
            }
            ::v-deep .content-body {
              display: flex;
              color: $text-normal;
              line-height: 22px;

              * {
                word-wrap: break-word;
                max-width: 225px;
              }

              a {
                color: $text-link;
              }

              strong {
                font-weight: 800;
              }

              em {
                font-style: italic;
              }

              blockquote {
                border-left: 3px $background-accent solid;
                padding-left: 3px;
              }

              .spoiler {
                background: $background-tertiary;
                padding: 0 2px;
                border-radius: 4px;
                cursor: pointer;

                span {
                  opacity: 0;
                }

                &.active {
                  background: $background-secondary;
                  cursor: default;
                  span {
                    opacity: 1;
                  }
                }
              }

              code {
                font-family: Consolas, Andale Mono WT, Andale Mono, Lucida Console, Lucida Sans Typewriter,
                  DejaVu Sans Mono, Bitstream Vera Sans Mono, Liberation Mono, Nimbus Mono L, Monaco, Courier New,
                  Courier, monospace;
                background: $background-secondary;
                border-radius: 3px;
                padding: 0 3px;
                font-size: 0.875rem;
                line-height: 1.125rem;
                text-indent: 0;
                white-space: pre-wrap;
              }

              div {
                flex: 1;

                pre {
                  color: $interactive-normal;
                  border: 1px solid $background-tertiary;
                  background: $background-secondary;
                  padding: 8px 6px;
                  margin: 4px 0;
                  border-radius: 4px;
                  display: block;
                  flex: 1;

                  code {
                    display: block;
                  }
                }
              }
            }
          }
        }

        &.event {
          flex: 1;
          overflow: hidden;
          display: flex;
          height: 15px;
          color: $text-muted;
          cursor: default;

          span {
            white-space: nowrap;
            max-width: 250px;
            text-overflow: ellipsis;
            overflow: hidden;
            height: 15px;

            strong {
              font-weight: 600;
            }

            i {
              float: right;
              font-style: italic;
              font-size: 10px;
            }
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

      .accent {
        width: 100%;
        height: 1px;
        background: rgba($color: #fff, $alpha: 0.05);
        margin: 5px 0 10px 0;
      }

      .text-container {
        flex: 1;
        width: 100%;
        height: 100%;
        background-color: rgba($color: #fff, $alpha: 0.05);
        border-radius: 5px;
        display: flex;

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
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import { formatRelative } from 'date-fns'

  import Markdown from './markdown'
  import Content from './context.vue'

  const length = 512 // max length of message

  @Component({
    name: 'neko-chat',
    components: {
      'neko-markdown': Markdown,
      'neko-context': Content,
    },
  })
  export default class extends Vue {
    @Ref('history') readonly _history!: HTMLElement
    @Ref('context') readonly _context!: any

    content = ''

    get id() {
      return this.$accessor.user.id
    }

    get muted() {
      return this.$accessor.user.muted
    }

    get history() {
      return this.$accessor.chat.history
    }

    @Watch('history')
    onHistroyChange() {
      this.$nextTick(() => {
        this._history.scrollTop = this._history.scrollHeight
      })
    }

    @Watch('muted')
    onMutedChange(muted: boolean) {
      if (muted) {
        this.content = ''
      }
    }

    mounted() {
      this.$nextTick(() => {
        this._history.scrollTop = this._history.scrollHeight
      })
    }

    member(id: string) {
      return this.$accessor.user.members[id]
    }

    timestamp(time: Date) {
      return formatRelative(time, new Date())
    }

    onContext(event: MouseEvent, data: any) {
      this._context.open(event, data)
    }

    onClick(event: { target?: HTMLElement; preventDefault(): void }) {
      const { target } = event
      if (!target) {
        return
      }

      if (target.tagName.toLowerCase() === 'span' && target.classList.contains('spoiler')) {
        target.classList.add('active')
        event.preventDefault()
      }

      if (!target.parentElement) {
        return
      }

      if (target.parentElement.tagName.toLowerCase() === 'span' && target.parentElement.classList.contains('spoiler')) {
        target.parentElement.classList.add('active')
        event.preventDefault()
      }
    }

    onKeyDown(event: KeyboardEvent) {
      if (this.muted) {
        return
      }

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

      this.$accessor.chat.sendMessage(this.content)

      this.content = ''
      event.preventDefault()
    }
  }
</script>
