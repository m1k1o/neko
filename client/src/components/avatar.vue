<template>
	<img :src="`https://ui-avatars.com/api/?name=${seed}&size=${size}`" />
</template>

<script lang="ts">
  import { Vue, Component, Prop } from 'vue-property-decorator'

  @Component({
    name: 'neko-avatar',
  })
  export default class extends Vue {
  	@Prop(String) readonly seed: string | undefined;
  	@Prop(Number) readonly size: number | undefined;
  }
</script>
