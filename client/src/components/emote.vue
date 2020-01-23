<template>
  <div ref="emote" @click.stop.prevent="run" class="emote">
    <i :class="classes"></i>
    <i :class="classes"></i>
    <i :class="classes"></i>
    <i :class="classes"></i>
    <i :class="classes"></i>
    <i :class="classes"></i>
    <i :class="classes"></i>
  </div>
</template>

<style lang="scss" scoped>
  .emote {
    width: 150px;
    height: 30px;
    position: absolute;
    bottom: 0;
    right: 0;

    i {
      position: absolute;
      width: 30px;
      height: 30px;
      color: #fff;
      font-size: 30px;
      line-height: 30px;
      text-align: center;

      &.heart {
        color: rgb(204, 72, 72);
      }
      &.poo {
        color: rgb(112, 89, 58);
      }
      &.grin {
        color: rgb(228, 194, 84);
      }
      &.dizzy {
        color: rgb(199, 199, 199);
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

      const emotes: any = {
        heart: 'fa-heart',
        poo: 'fa-poo',
        grin: 'fa-grin-tears',
        ghost: 'fa-ghost',
      }

      this.classes = ['fas', emotes[this.emote.type] || 'fa-heart', this.emote.type]

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
