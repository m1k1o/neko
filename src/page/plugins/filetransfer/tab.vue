<template>
  <span>
    <li :class="{ active: tab === 'filetransfer' }" @click.prevent="emit('tab', 'filetransfer')">
      <i class="fas fa-file"></i>
      <span v-show="tab === 'filetransfer'">Filetransfer</span>
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
