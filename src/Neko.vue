<template>
  <div>
    <h1>Hello world, Neko.</h1>
    <button @click="visible = !visible">Toggle</button>
    <div v-if="visible">Is this visible?</div>
  </div>
</template>

<script lang="ts">
  import { Vue, Component, Ref } from 'vue-property-decorator'

  @Component({
    name: 'neko',
  })
  export default class extends Vue {
    public visible: boolean = false
  }
</script>
