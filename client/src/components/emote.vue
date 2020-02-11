<template>
  <div ref="emote" @click.stop.prevent="run" class="emote-container">
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
  .emote-container {
    width: 25%;
    height: 10%;
    position: absolute;
    bottom: 0;
    right: 0;

    .emote {
      position: absolute;
      width: 100px;
      height: 100%;
      color: #fff;
      transform: scale(0.7);
      text-align: center;
      background-size: contain;
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
      const range = 50
      let count = 0
      let finish: Array<Promise<any>> = []

      this.classes = ['emote', this.emote.type]

      for (let child of this.container.children) {
        const ele = child as HTMLElement
        ele.style['left'] = `${count % 2 ? this.$anime.random(0, range) : this.$anime.random(-range, 0)}%`
        ele.style['opacity'] = `0`

        const animation = this.$anime({
          targets: child,
          keyframes: [
            { left: `${count % 2 ? this.$anime.random(0, range) : this.$anime.random(-range, 0)}%`, opacity: 1 },
            { left: `${count % 2 ? this.$anime.random(-range, 0) : this.$anime.random(0, range)}%`, opacity: 0.5 },
            { left: `${count % 2 ? this.$anime.random(0, range) : this.$anime.random(-range, 0)}%`, opacity: 0 },
          ],
          elasticity: 600,
          rotate: this.$anime.random(-35, 35),
          top: `${this.$anime.random(-200, -600)}%`,
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
