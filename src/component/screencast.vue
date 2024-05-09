<template>
  <img :src="imageSrc" @load="onImageLoad" @error="onImageError" />
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import type { RoomApi } from './api'

const REFRESH_RATE = 1e3
const ERROR_DELAY_MS = 2500

const imageSrc = ref('')

const props = defineProps<{
  image: string
  enabled: boolean
  api: RoomApi
}>()

const emit = defineEmits(['imageReady'])

watch(() => props.image, (image) => {
  imageSrc.value = image
})

let isRunning = false
let isStopped = false

async function loop() {
  if (isRunning) return
  isRunning = true

  while (!isStopped) {
    const lastLoad = Date.now()

    try {
      const res = await props.api.screenCastImage({ responseType: 'blob' })
      imageSrc.value = URL.createObjectURL(res.data)

      const delay = lastLoad - Date.now() + REFRESH_RATE
      if (delay > 0) {
        await new Promise((res) => setTimeout(res, delay))
      }
    } catch {
      await new Promise((res) => setTimeout(res, ERROR_DELAY_MS))
    }
  }

  isRunning = false
  imageSrc.value = ''
}

onMounted(() => {
  if (props.enabled) {
    start()
  }
})

onBeforeUnmount(() => {
  stop()
})

function start() {
  isStopped = false

  if (!isRunning) {
    setTimeout(loop, 0)
  }
}

function stop() {
  isStopped = true
}

function onEnabledChanged(enabled: boolean) {
  if (enabled) {
    start()
  } else {
    stop()
  }
}

watch(() => props.enabled, onEnabledChanged)

function onImageLoad() {
  URL.revokeObjectURL(imageSrc.value)
  emit('imageReady', isRunning)
}

function onImageError() {
  if (imageSrc.value) URL.revokeObjectURL(imageSrc.value)
  emit('imageReady', false)
}
</script>
