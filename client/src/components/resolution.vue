<template>
  <vue-context class="context" ref="context">
    <template v-for="(conf, i) in configurations">
      <li
        :key="i"
        @click="screenSet(conf)"
        :class="[conf.width === width && conf.height === height && conf.rate === rate ? 'active' : '']"
      >
        <i class="fas fa-desktop"></i>
        <span>{{ conf.width }}x{{ conf.height }}</span>
        <small>{{ conf.rate }}</small>
      </li>
    </template>
    <li class="fit-row" @click="fitToWindow">
      <i class="fas fa-expand-arrows-alt"></i>
      <span>适应窗口大小</span>
      <small>{{ windowSize }}</small>
    </li>
    <li class="custom-row" @click.stop>
      <i class="fas fa-pen"></i>
      <input
        v-model="customWidth"
        class="custom-input"
        type="number"
        min="320"
        max="7680"
        placeholder="宽"
        @keydown.stop
        @click.stop
      />
      <span class="custom-sep">x</span>
      <input
        v-model="customHeight"
        class="custom-input"
        type="number"
        min="240"
        max="4320"
        placeholder="高"
        @keydown.stop
        @click.stop
      />
      <span class="custom-sep">@</span>
      <input
        v-model="customRate"
        class="custom-input custom-input-rate"
        type="number"
        min="1"
        max="240"
        placeholder="帧率"
        @keydown.stop
        @click.stop
      />
      <button class="custom-btn" @click.stop="applyCustom">✓</button>
    </li>
  </vue-context>
</template>

<style lang="scss" scoped>
  .context {
    background-color: $background-floating;
    background-clip: padding-box;
    border-radius: 0.25rem;
    display: block;
    margin: 0;
    padding: 5px;
    min-width: 150px;
    z-index: 1500;
    position: fixed;
    list-style: none;
    box-sizing: border-box;
    max-height: calc(100% - 50px);
    overflow-y: auto;
    color: $interactive-normal;
    user-select: none;
    box-shadow: $elevation-high;
    scrollbar-width: thin;
    scrollbar-color: $background-secondary transparent;

    &::-webkit-scrollbar { width: 8px; }
    &::-webkit-scrollbar-track { background-color: transparent; }
    &::-webkit-scrollbar-thumb {
      background-color: $background-secondary;
      border: 2px solid $background-floating;
      border-radius: 4px;
    }
    &::-webkit-scrollbar-thumb:hover { background-color: $background-floating; }

    > li {
      margin: 0;
      position: relative;
      align-content: center;
      display: flex;
      flex-direction: row;
      padding: 8px;
      cursor: pointer;
      border-radius: 3px;

      i { margin-right: 10px; }
      span { flex-grow: 1; }
      small {
        font-size: 0.7em;
        justify-self: flex-end;
        align-self: flex-end;
      }

      &.active,
      &:hover,
      &:focus {
        text-decoration: none;
        background-color: $background-modifier-hover;
        color: $interactive-hover;
      }

      &:focus { outline: 0; }
    }

    &:focus { outline: 0; }

    .fit-row {
      border-top: 1px solid rgba(255, 255, 255, 0.1);
      margin-top: 4px;
      padding-top: 8px;
    }

    .custom-row {
      border-top: 1px solid rgba(255, 255, 255, 0.1);
      margin-top: 4px;
      padding-top: 8px;
      cursor: default;
      gap: 4px;
      align-items: center;

      &:hover { background-color: transparent; }

      i { flex-shrink: 0; }
    }

    .custom-sep {
      flex-grow: 0;
      color: rgba(255, 255, 255, 0.4);
      font-size: 0.85em;
    }

    .custom-input {
      width: 52px;
      background: rgba(255, 255, 255, 0.1);
      border: 1px solid rgba(255, 255, 255, 0.2);
      border-radius: 3px;
      color: inherit;
      font-size: 0.8em;
      padding: 2px 4px;
      outline: none;
      text-align: center;
      -moz-appearance: textfield;

      &::-webkit-outer-spin-button,
      &::-webkit-inner-spin-button { -webkit-appearance: none; }

      &:focus { border-color: rgba(255, 255, 255, 0.5); }
    }

    .custom-input-rate { width: 38px; }

    .custom-btn {
      background: rgba(114, 137, 218, 0.6);
      border: none;
      border-radius: 3px;
      color: #fff;
      cursor: pointer;
      font-size: 0.85em;
      padding: 2px 6px;
      flex-shrink: 0;

      &:hover { background: rgba(114, 137, 218, 0.9); }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'
  import { ScreenResolution } from '~/neko/types'

  // @ts-ignore
  import { VueContext } from 'vue-context'

  @Component({
    name: 'neko-resolution',
    components: {
      'vue-context': VueContext,
    },
  })
  export default class extends Vue {
    @Ref('context') readonly context!: VueContext

    public customWidth: number | string = ''
    public customHeight: number | string = ''
    public customRate: number | string = ''

    get width() { return this.$accessor.video.width }
    get height() { return this.$accessor.video.height }
    get rate() { return this.$accessor.video.rate }
    get configurations() { return this.$accessor.video.configurations }

    get windowSize() {
      return `${Math.floor(window.innerWidth / 2) * 2}x${Math.floor(window.innerHeight / 2) * 2}`
    }

    open(event: MouseEvent) {
      // Pre-fill with current resolution
      this.customWidth = this.width
      this.customHeight = this.height
      this.customRate = this.rate
      this.context.open(event)
    }

    fitToWindow() {
      const w = Math.floor(window.innerWidth / 2) * 2   // ensure even number
      const h = Math.floor(window.innerHeight / 2) * 2
      this.$accessor.video.screenSet({ width: w, height: h, rate: this.rate })
      this.context.close()
    }

    screenSet(resolution: ScreenResolution) {
      this.$accessor.video.screenSet(resolution)
    }

    applyCustom() {
      const w = parseInt(String(this.customWidth), 10)
      const h = parseInt(String(this.customHeight), 10)
      const r = parseInt(String(this.customRate), 10)
      if (!w || !h || !r || w < 320 || h < 240 || r < 1) return
      this.$accessor.video.screenSet({ width: w, height: h, rate: r })
      this.context.close()
    }
  }
</script>
