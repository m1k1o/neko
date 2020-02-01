<template>
  <div class="emotes">
    <ul v-if="!muted">
      <li><div @click.stop.prevent="click('heart')" class="heart" /></li>
      <li><div @click.stop.prevent="click('sleep')" class="sleep" /></li>
      <li><div @click.stop.prevent="click('laughing')" class="laughing" /></li>
      <li><div @click.stop.prevent="click('celebrate')" class="celebrate" /></li>
      <li><div @click.stop.prevent="click('exclam')" class="exclam" /></li>
      <li><div @click.stop.prevent="click('clap')" class="clap" /></li>
    </ul>
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

        div {
          width: 24px;
          height: 24px;
          cursor: pointer;
          background-size: contain;

          &.celebrate {
            background-image: url('../assets/images/emote/celebrate.png');
          }

          &.clap {
            background-image: url('../assets/images/emote/clap.png');
          }

          &.exclam {
            background-image: url('../assets/images/emote/exclam.png');
          }

          &.heart {
            background-image: url('../assets/images/emote/heart.png');
          }

          &.laughing {
            background-image: url('../assets/images/emote/laughing.png');
          }

          &.sleep {
            background-image: url('../assets/images/emote/sleep.png');
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component } from 'vue-property-decorator'

  @Component({
    name: 'neko-emotes',
  })
  export default class extends Vue {
    get muted() {
      return this.$accessor.user.muted
    }

    click(emote: string) {
      this.$accessor.chat.sendEmote(emote)
    }
  }
</script>
