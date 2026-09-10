<template>
  <!--
      <img :src="`https://ui-avatars.com/api/?name=${seed}&size=${size}`" />
  -->
  <div
    class="avatar"
    :class="{ 'has-image': avatar && !imageFailed }"
    :style="{
      width: size + 'px',
      height: size + 'px',
      lineHeight: size + 'px',
      fontSize: size / 2 + 'px',
      backgroundColor: Background(seed),
    }"
  >
    <img v-if="avatar && !imageFailed" :src="avatar" alt="" @error="imageFailed = true" />
    <span v-else>{{ initials }}</span>
  </div>
</template>

<style lang="scss" scoped>
  .avatar {
    user-select: none;
    text-align: center;
    background: white;
    color: black;
    display: inline-block;
    overflow: hidden;
    border-radius: 50%;

    img {
      display: block;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    span {
      display: block;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop, Watch } from 'vue-property-decorator'

  @Component({
    name: 'neko-avatar',
  })
  export default class extends Vue {
    @Prop({ type: String, default: '' }) readonly seed!: string
    @Prop({ type: String, default: '' }) readonly avatar!: string
    @Prop({ type: Number, default: 40 }) readonly size!: number

    imageFailed = false

    get initials() {
      return this.seed.substring(0, 2).toUpperCase()
    }

    @Watch('avatar')
    onAvatarChanged() {
      this.imageFailed = false
    }

    Background(seed: string) {
      let a = 0,
        b = 0,
        c = 0

      for (let i = 0; i < seed.length; i++) {
        a += seed.charCodeAt(i) * 3
        b += seed.charCodeAt(i) * 5
        c += seed.charCodeAt(i) * 7
      }

      let x = Math.floor(128 + (a % 128))
      let y = Math.floor(128 + (b % 128))
      let z = Math.floor(128 + (c % 128))
      return 'rgb(' + x + ',' + y + ',' + z + ')'
    }
  }
</script>
