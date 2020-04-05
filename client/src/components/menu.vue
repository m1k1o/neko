<template>
  <ul>
    <li><i @click.stop.prevent="about" class="fas fa-question-circle" /></li>
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
  </ul>
</template>

<style lang="scss" scoped>
  ul {
    li {
      display: inline-block;
      margin-right: 10px;

      i {
        font-size: 24px;
        cursor: pointer;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-menu' })
  export default class extends Vue {
    get admin() {
      return this.$accessor.user.admin
    }

    about() {
      this.$accessor.client.toggleAbout()
    }
  }
</script>
