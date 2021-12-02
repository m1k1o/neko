<template>
  <div class="chat">
    <ul class="chat-history" ref="history" @click="onClick">
      <template v-for="(message, index) in history">
        <li
          :key="index"
          class="message"
          v-if="message.type === 'text'"
          :class="{
            bulk: index > 0 && history[index - 1].id == message.id && history[index - 1].type === 'text',
          }"
        >
          <div class="author" @contextmenu.stop.prevent="onContext($event, { member: member(message.id) })">
            <neko-avatar class="avatar" :seed="member(message.id).displayname" :size="40" />
          </div>
          <div class="content">
            <div class="content-head">
              <span>{{ member(message.id).displayname }}</span>
              <span class="timestamp">{{ timestamp(message.created) }}</span>
            </div>
            <neko-markdown class="content-body" :source="message.content" />
          </div>
        </li>
        <li :key="index" class="event" v-if="message.type === 'event'">
          <div
            class="content"
            v-tooltip="{
              content: timestamp(message.created),
              placement: 'left',
              offset: 3,
              boundariesElement: 'body',
            }"
          >
            <strong v-if="message.id === id && $te('you')">{{ $t('you') }}</strong>
            <strong v-else>{{ member(message.id).displayname }}</strong>
            {{ message.content }}
          </div>
        </li>
      </template>
    </ul>
    <neko-context ref="context" />
    <div v-if="!muted" class="chat-send">
      <div class="accent" />
      <div class="text-container">
        <textarea ref="input" :placeholder="$t('send_a_message')" @keydown="onKeyDown" v-model="content" />
        <neko-emoji v-if="emoji" @picked="onEmojiPicked" @done="emoji = false" />
        <i class="emoji-menu fas fa-laugh" @click.stop.prevent="onEmoji"></i>
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
        padding: 10px 5px 0px 10px;
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
        overflow: hidden;
        user-select: text;
        word-wrap: break-word;

        &.message {
          padding-top: 15px;
          font-size: 16px;

          .author {
            flex-grow: 0;
            flex-shrink: 0;
            overflow: hidden;
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: $style-primary;
            margin-right: 10px;

            .avatar {
              width: 100%;
            }
          }

          .content {
            flex: 1;
            display: flex;
            flex-direction: column;
            box-sizing: border-box;
            word-wrap: break-word;
            min-width: 0;

            .content-head {
              cursor: default;
              width: 100%;
              margin-bottom: 3px;
              display: block;

              span {
                display: inline-block;
                color: $text-normal;
                font-weight: 500;
              }

              .timestamp {
                color: $text-muted;
                font-size: 0.7rem;
                font-weight: 500;
                margin-left: 0.3rem;
                line-height: 12px;

                &::first-letter {
                  text-transform: uppercase;
                }
              }
            }

            ::v-deep .content-body {
              color: $text-normal;
              line-height: 22px;
              word-wrap: break-word;
              overflow-wrap: break-word;

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

              span {
                &.spoiler {
                  background: $background-tertiary;
                  padding: 0 2px;
                  border-radius: 4px;
                  cursor: pointer;

                  span {
                    opacity: 0;
                  }
                }

                &.spoiler.active {
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

              pre {
                flex: 1;
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

          &.bulk {
            padding-top: 0px;

            .author {
              visibility: hidden;
              height: 0;
            }

            .content-head {
              display: none;
            }
          }
        }

        &.event {
          color: $text-muted;
          cursor: default;

          .content {
            min-width: 0;
            box-sizing: border-box;
            word-wrap: break-word;
            display: inline-block;
            vertical-align: baseline;
            line-height: 20px;

            strong {
              font-weight: 600;
            }

            i {
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
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import { formatRelative } from 'date-fns'

  import { Member } from '~/neko/types'

  import Markdown from './markdown'
  import Content from './context.vue'
  import Emoji from './emoji.vue'
  import Avatar from './avatar.vue'

  const length = 512 // max length of message

  @Component({
    name: 'neko-chat',
    components: {
      'neko-markdown': Markdown,
      'neko-context': Content,
      'neko-emoji': Emoji,
      'neko-avatar': Avatar,
    },
  })
  export default class extends Vue {
    @Ref('input') readonly _input!: HTMLTextAreaElement
    @Ref('history') readonly _history!: HTMLElement
    @Ref('context') readonly _context!: any

    emoji = false
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
      return this.$accessor.user.members[id] || { id, displayname: this.$t('somebody') }
    }

    timestamp(time: Date) {
      const str = formatRelative(time, new Date())
      return `${str.charAt(0).toUpperCase()}${str.slice(1)}`
    }

    onEmoji() {
      this.emoji = !this.emoji
      this._input.focus()
    }

    onEmojiPicked(emoji: string) {
      const text = `:${emoji}:`
      if (this._input.selectionStart || this._input.selectionStart === 0) {
        var startPos = this._input.selectionStart
        var endPos = this._input.selectionEnd
        this.content = this.content.substring(0, startPos) + text + this.content.substring(endPos, this.content.length)
        this.$nextTick(() => {
          this._input.selectionStart = startPos + text.length
          this._input.selectionEnd = startPos + text.length
        })
      } else {
        this.content += text
      }
      this._input.focus()
      this.emoji = false
    }

    onContext(event: MouseEvent, { member }: { member: Member }) {
      if (member.id === this.id) {
        return
      }
      this._context.open(event, { member })
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
