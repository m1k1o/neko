<template>
  <span>
    <li :class="{ active: tab === 'chat' }" @click.prevent="emit('tab', 'chat')">
      <i class="fas fa-comment-alt"></i>
      <span v-show="tab === 'chat'">Chat</span>
    </li>
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  tab: string
}>()

const emit = defineEmits<{
  (event: 'tab', tab: string): void
}>()
</script>
