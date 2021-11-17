<template>
  <div class="emotes" @mouseleave="stopSendingEmotes" @mouseup="stopSendingEmotes">
    <ul v-if="!muted">
      <li v-for="emote in recent" :key="emote">
        <div :class="['emote', emote]" @mousedown.stop.prevent="startSendingEmotes(emote)" />
      </li>
      <li>
        <i @click.stop.prevent="open" class="fas fa-grin-beam"></i>
      </li>
    </ul>
    <vue-context class="context" ref="context">
      <li v-for="emote in emotes" :key="emote">
        <div @click="sendEmote(emote)" :class="['emote', emote]" />
      </li>
    </vue-context>
  </div>
</template>

<style lang="scss" scoped>
  .emotes {
    ul {
      display: flex;
      flex-direction: row;
      justify-content: center;
      align-items: center;

      li {
        font-size: 24px;
        margin: 0 5px;

        i,
        div {
          cursor: pointer;
        }
      }
    }

    .context {
      background-color: $background-floating;
      background-clip: padding-box;
      border-radius: 0.25rem;
      display: flex;
      margin: 0;
      padding: 5px;
      width: 220px;
      z-index: 1500;
      position: fixed;
      list-style: none;
      box-sizing: border-box;
      max-height: calc(100% - 50px);
      color: $interactive-normal;
      flex-wrap: wrap;
      user-select: none;
      box-shadow: $elevation-high;

      > li {
        margin: 0;
        position: relative;
        align-content: center;
        padding: 5px;
        border-radius: 3px;

        .emote {
          width: 24px;
          height: 24px;
        }

        &:hover,
        &:focus {
          text-decoration: none;
          background-color: $background-modifier-hover;
          color: $interactive-hover;
        }

        &:focus {
          outline: 0;
        }
      }

      &:focus {
        outline: 0;
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Ref, Component } from 'vue-property-decorator'
  import { get, set } from '../utils/localstorage'

  // @ts-ignore
  import { VueContext } from 'vue-context'

  @Component({
    name: 'neko-emotes',
    components: {
      'vue-context': VueContext,
    },
  })
  export default class extends Vue {
    @Ref('context') readonly context!: any
    recent: string[] = JSON.parse(get('emote_recent', '[]'))

    get emotes() {
      return [
        'anger',
        'bomb',
        'sleep',
        'explode',
        'sweat',
        'poo',
        'hundred',
        'alert',
        'punch',
        'wave',
        'okay',
        'thumbs-up',
        'clap',
        'prey',
        'celebrate',
        'flame',
        'goof',
        'love',
        'cool',
        'smerk',
        'worry',
        'ouch',
        'cry',
        'surprised',
        'quiet',
        'rage',
        'annoy',
        'steamed',
        'scared',
        'terrified',
        'sleepy',
        'dead',
        'happy',
        'roll-eyes',
        'thinking',
        'clown',
        'sick',
        'rofl',
        'drule',
        'sniff',
        'sus',
        'party',
        'odd',
        'hot',
        'cold',
        'blush',
        'sad',
      ].filter((v) => !this.recent.includes(v))
    }

    get muted() {
      return this.$accessor.user.muted
    }

    open(event: MouseEvent) {
      this.context.open(event)
    }

    sendEmote(emote: string) {
      if (!this.recent.includes(emote)) {
        if (this.recent.length > 4) {
          this.recent.shift()
        }
        this.recent.push(emote)
        set('emote_recent', JSON.stringify(this.recent))
      }
      this.$accessor.chat.sendEmote(emote)
    }

    private interval!: number

    startSendingEmotes(emote: string) {
      this.$accessor.chat.sendEmote(emote)
      this.stopSendingEmotes()

      this.interval = window.setInterval(() => {
        this.$accessor.chat.sendEmote(emote)
      }, 350)
    }

    stopSendingEmotes() {
      if (this.interval) {
        clearInterval(this.interval)
      }
    }
  }
</script>
