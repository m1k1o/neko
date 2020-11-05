<template>
  <div>
    <h1>Hello world, Neko.</h1>
  </div>
</template>

<script lang="ts">
  import { Vue, Component, Ref } from 'vue-property-decorator'

  @Component({
    name: 'neko',
  })
  export default class extends Vue {}
</script>
