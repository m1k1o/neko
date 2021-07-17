<template>
  <div class="clipboard" v-if="opened" @click="$event.stopPropagation()">
    <textarea ref="textarea" v-model="clipboard" @focus="$event.target.select()" />
  </div>
</template>

<style lang="scss" scoped>
  .clipboard {
    background-color: $background-primary;
    border-radius: 0.25rem;
    display: block;
    padding: 5px;

    position: absolute;
    bottom: 10px;
    right: 10px;

    &,
    textarea {
      max-width: 320px;
      width: 100%;
      max-height: 120px;
      height: 100%;
    }

    textarea {
      border: 0;
      color: $text-normal;
      background: none;

      &::selection {
        background: $text-normal;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'

  @Component({
    name: 'neko-clipboard',
  })
  export default class extends Vue {
    @Ref('textarea') readonly _textarea!: HTMLTextAreaElement

    private opened: boolean = false
    private typing?: number

    get clipboard() {
      return this.$accessor.remote.clipboard
    }

    set clipboard(data: string) {
      this.$accessor.remote.setClipboard(data)

      if (this.typing) {
        clearTimeout(this.typing)
        this.typing = undefined
      }

      this.typing = window.setTimeout(() => this.$accessor.remote.sendClipboard(this.clipboard), 500)
    }

    open() {
      this.opened = true
      document.body.addEventListener('click', this.close)
      window.setTimeout(() => this._textarea.focus(), 0)
    }

    close() {
      this.opened = false
      document.body.removeEventListener('click', this.close)
    }
  }
</script>
