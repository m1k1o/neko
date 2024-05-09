<template>
  <div class="header">
    <div class="neko">
      <span class="logo"><b>n</b>.eko</span>
      <div class="server">
        <span>Server:</span>
        <input type="text" placeholder="URL" v-model="url" />
        <button @click="setUrl">change</button>
      </div>
      <ul class="menu">
        <li>
          <i class="fas fa-bars toggle" @click="toggleMenu" />
        </li>
      </ul>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  @import '../assets/styles/_variables.scss';

  .header {
    flex: 1;
    display: flex;
    flex-direction: row;
    align-items: center;
    height: 100%;

    .neko {
      flex: 1;
      display: flex;
      justify-content: space-between;
      align-items: center;
      width: 150px;
      margin-left: 20px;

      .logo {
        font-size: 30px;
        line-height: 30px;

        b {
          font-weight: 900;
        }
      }

      .server {
        max-width: 850px;
        width: 100%;
        margin: 0 20px;
        display: flex;
        align-items: center;

        input {
          margin: 0 5px;
          width: 100%;
        }
      }
    }

    .menu {
      justify-self: flex-end;
      margin-right: 10px;
      white-space: nowrap;

      li {
        display: inline-block;
        margin-right: 10px;

        i {
          display: block;
          width: 30px;
          height: 30px;
          text-align: center;
          line-height: 32px;
          border-radius: 3px;
          cursor: pointer;
        }

        .toggle {
          background: $background-primary;
        }
      }
    }
  }
</style>

<script lang="ts" setup>
import { ref, watch } from 'vue'
import type Neko from '@/component/main.vue'

const props = defineProps<{
  neko: typeof Neko
}>()

const url = ref('')

watch(() => props.neko.state.connection.url, (u) => {
  url.value = u
})

async function setUrl() {
  if (url.value == '') {
    url.value = location.href
  }

  await props.neko.setUrl(url.value)
}

function toggleMenu() {
  props.neko.toggleSide()
}
</script>
