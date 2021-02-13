<template>
  <!--
      <img :src="`https://ui-avatars.com/api/?name=${seed}&size=${size}`" />
  -->
  <div
    class="avatar"
    :style="{
      width: size + 'px',
      height: size + 'px',
      lineHeight: size + 'px',
      fontSize: size / 2 + 'px',
      backgroundColor: Background(seed),
    }"
  >
    {{ seed.substring(0, 2).toUpperCase() }}
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
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop } from 'vue-property-decorator'

  @Component({
    name: 'neko-avatar',
  })
  export default class extends Vue {
    @Prop(String) readonly seed: string | undefined
    @Prop(Number) readonly size: number | undefined

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
