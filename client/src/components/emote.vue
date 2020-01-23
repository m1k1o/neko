<template>
  <div ref="emote" @click.stop.prevent="run" class="emote">
    <div :class="classes" />
    <div :class="classes" />
    <div :class="classes" />
    <div :class="classes" />
    <div :class="classes" />
    <div :class="classes" />
    <div :class="classes" />
  </div>
</template>

<style lang="scss" scoped>
  .emote {
    width: 150px;
    height: 30px;
    position: absolute;
    bottom: 0;
    right: 0;

    div {
      position: absolute;
      width: 30px;
      height: 30px;
      color: #fff;
      font-size: 30px;
      line-height: 30px;
      text-align: center;
      background-size: contain;

      &.celebrate {
        background-image: url('../assets/celebrate.png');
      }

      &.clap {
        background-image: url('../assets/clap.png');
      }

      &.exclam {
        background-image: url('../assets/exclam.png');
      }

      &.heart {
        background-image: url('../assets/heart.png');
      }

      &.laughing {
        background-image: url('../assets/laughing.png');
      }

      &.sleep {
        background-image: url('../assets/sleep.png');
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue, Prop } from 'vue-property-decorator'

  @Component({ name: 'neko-emote' })
  export default class extends Vue {
    @Prop({
      required: true,
    })
    id!: string

    @Ref('emote') container!: HTMLElement

    get emote() {
      return this.$accessor.chat.emotes[this.id]
    }

    private classes: string[] = []

    mounted() {
      const range = 90
      let count = 0
      let finish: Array<Promise<any>> = []

      this.classes = [this.emote.type]

      for (let child of this.container.children) {
        const ele = child as HTMLElement
        ele.style['left'] = `${count % 2 ? this.$anime.random(0, range) : this.$anime.random(-range, 0)}px`
        ele.style['opacity'] = `0`

        const animation = this.$anime({
          targets: child,
          keyframes: [
            { left: count % 2 ? this.$anime.random(0, range) : this.$anime.random(-range, 0), opacity: 1 },
            { left: count % 2 ? this.$anime.random(-range, 0) : this.$anime.random(0, range), opacity: 0.5 },
            { left: count % 2 ? this.$anime.random(0, range) : this.$anime.random(-range, 0), opacity: 0 },
          ],
          elasticity: 600,
          rotate: this.$anime.random(-35, 35),
          top: this.$anime.random(-100, -250),
          duration: this.$anime.random(1000, 2000),
          easing: 'easeInOutQuad',
        })

        count++
        finish.push(animation.finished)
      }

      Promise.all(finish).then(() => {
        this.$emit('done', this.id)
        this.$accessor.chat.delEmote(this.id)
      })
    }
  }
</script>
