<template>
  <ul class="menu-actions">
    <li>
      <button type="button" class="action-button" aria-label="About n.eko" @click.stop.prevent="about">
        <i class="fas fa-question-circle" aria-hidden="true" />
      </button>
    </li>
    <li>
      <i
        class="fas fa-shield-alt"
        v-tooltip="{
          content: $t('admin_loggedin'),
          placement: 'right',
          offset: 5,
          boundariesElement: 'body',
        }"
        v-if="admin"
      />
    </li>
    <li class="locale-picker">
      <label>
        <span class="sr-only">Language</span>
        <select v-model="$i18n.locale">
          <option v-for="(lang, i) in langs" :key="`Lang${i}`" :value="lang">
            {{ lang }}
          </option>
        </select>
      </label>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  .menu-actions {
    display: flex;
    align-items: center;
    gap: 6px;

    li {
      display: flex;
      align-items: center;

      .action-button {
        width: 34px;
        height: 34px;
        display: grid;
        place-items: center;
        border: 1px solid transparent;
        border-radius: 9px;
        color: $interactive-normal;
        background: transparent;
        cursor: pointer;
        transition: color 0.18s ease, background 0.18s ease, border-color 0.18s ease;

        &:hover,
        &:focus-visible {
          color: $interactive-hover;
          background: $background-modifier-hover;
          border-color: rgba($text-normal, 0.12);
        }

        i {
          font-size: 16px;
        }
      }

      > i {
        color: $style-primary;
        font-size: 15px;
      }
    }
  }

  .locale-picker select {
    appearance: none;
    background: $background-tertiary;
    border: 1px solid rgba($text-normal, 0.12);
    color: $interactive-normal;
    cursor: pointer;
    border-radius: 8px;
    height: 34px;
    min-width: 58px;
    padding: 0 26px 0 10px;
    display: block;
    font-size: 12px;
    font-weight: 600;

    option {
      font-weight: normal;
      color: $text-normal;
      background-color: $background-tertiary;
    }

    &:hover {
      border-color: rgba($text-normal, 0.24);
    }
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>

<script lang="ts">
  import { Component, Vue, Watch } from 'vue-property-decorator'
  import { messages } from '~/locale'
  import { set } from '~/utils/localstorage'

  @Component({ name: 'neko-menu' })
  export default class extends Vue {
    get admin() {
      return this.$accessor.user.admin
    }

    get langs() {
      return Object.keys(messages)
    }

    about() {
      this.$accessor.client.toggleAbout()
    }

    @Watch('$i18n.locale')
    onLanguageChange(newLang: string) {
      set('lang', newLang)
    }

    mounted() {
      const default_lang = new URL(location.href).searchParams.get('lang')
      if (default_lang && this.langs.includes(default_lang)) {
        this.$i18n.locale = default_lang
      }
      const show_side = new URL(location.href).searchParams.get('show_side')
      if (show_side !== null) {
        this.$accessor.client.setSide(show_side === '1')
      }
      const mute_chat = new URL(location.href).searchParams.get('mute_chat')
      if (mute_chat !== null) {
        this.$accessor.settings.setSound(mute_chat !== '1')
      }
    }
  }
</script>
