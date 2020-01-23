<template>
  <div class="emotes">
    <ul v-if="!muted">
      <li><i @click.stop.prevent="click('heart')" class="fas fa-heart"></i></li>
      <li><i @click.stop.prevent="click('poo')" class="fas fa-poo"></i></li>
      <li><i @click.stop.prevent="click('grin')" class="fas fa-grin-tears"></i></li>
      <li><i @click.stop.prevent="click('ghost')" class="fas fa-ghost"></i></li>
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

        i {
          cursor: pointer;
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
