<template>
  <div class="about" @click="toggle">
    <div class="window">
      <div class="loading" v-if="loading">
        <div class="logo">
          <img src="@/assets/images/logo.svg" alt="n.eko" />
          <span><b>N</b>.EKO</span>
        </div>
        <div class="loader">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
        </div>
      </div>

      <div class="markdown-body" v-if="!loading" v-html="about"></div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .about {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba($color: $background-floating, $alpha: 0.8);

    display: flex;
    justify-content: center;
    align-items: center;

    .window {
      max-width: 70vw;
      background: $background-secondary;
      border-radius: 5px;
      max-height: 70vh;
      overflow-y: auto;
      overflow-x: hidden;

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

      .loading {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;

        .logo {
          display: flex;
          flex-direction: row;
          justify-content: center;
          align-items: center;
          margin: 40px 80px 0 80px;

          img {
            height: 90px;
            margin-right: 10px;
          }

          span {
            font-size: 30px;
            line-height: 56px;

            b {
              font-weight: 900;
            }
          }
        }

        .loader {
          width: 90px;
          height: 90px;
          position: relative;
          margin: 0 auto 20px auto;

          .bounce1,
          .bounce2 {
            width: 100%;
            height: 100%;
            border-radius: 50%;
            background-color: $style-primary;
            opacity: 0.6;
            position: absolute;
            top: 0;
            left: 0;

            -webkit-animation: bounce 2s infinite ease-in-out;
            animation: bounce 2s infinite ease-in-out;
          }

          .bounce2 {
            -webkit-animation-delay: -1s;
            animation-delay: -1s;
          }
        }
      }

      .markdown-body {
        margin: 50px 200px;
      }
    }
  }

  @keyframes bounce {
    0%,
    100% {
      transform: scale(0);
      -webkit-transform: scale(0);
    }
    50% {
      transform: scale(1);
      -webkit-transform: scale(1);
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import md, { HtmlOutputRule } from 'simple-markdown'

  @Component({ name: 'neko-about' })
  export default class extends Vue {
    loading = false

    get about() {
      return this.$accessor.client.about_page
    }

    mounted() {
      if (this.about === '') {
        this.loading = true
        this.$http
          .get<string>('https://raw.githubusercontent.com/nurdism/neko/master/docs/README.md')
          .then((res) => {
            return this.$http.post('https://api.github.com/markdown', {
              text: res.data,
              mode: 'gfm',
              context: 'github/gollum',
            })
          })
          .then((res) => {
            this.$accessor.client.setAbout(res.data)
            this.loading = false
          })
          .catch((err) => console.error(err))
      }
    }

    toggle(event: { target?: HTMLElement }) {
      if (event.target && event.target.classList.contains('about')) {
        this.$accessor.client.toggleAbout()
      }
    }
  }
</script>
