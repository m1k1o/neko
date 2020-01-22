<template>
  <div class="emoji">
    <ul>
      <li><i @click.stop.prevent="click('heart')" class="fas fa-heart"></i></li>
      <li><i @click.stop.prevent="click('poo')" class="fas fa-poo"></i></li>
      <li><i @click.stop.prevent="click('grin')" class="fas fa-grin-tears"></i></li>
      <li><i @click.stop.prevent="click('dizzy')" class="fas fa-dizzy"></i></li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  .emoji {
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
    name: 'neko-emoji',
  })
  export default class extends Vue {
    click(emoji: string) {
      this.$swal({
        title: 'Error!',
        text: 'This feature is not available yet',
        icon: 'error',
        confirmButtonText: 'Cool',
      })
    }
  }
</script>
