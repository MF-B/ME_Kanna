<template>
  <div class="item-icon-wrapper">
    <img 
      v-if="itemId && !isError"
      :src="iconUrl" 
      :alt="itemId"
      class="item-img"
      @error="handleError"
    />
    <div v-else class="fallback-icon">
      {{ getInitials(itemId) }}
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  itemId: { type: String, required: false }
})

const isError = ref(false)

// ID 变了重置错误状态
watch(() => props.itemId, () => {
  isError.value = false
})

// 拼接后端 API 地址 (假设后端在 8080 端口)
const iconUrl = computed(() => {
  if (!props.itemId) return ''
  const host = window.location.hostname
  return `http://${host}:8080/icon/${props.itemId}`
})

const handleError = () => {
  isError.value = true
}

const getInitials = (id) => {
  if (!id) return '?'
  const name = id.split(':')[1] || id
  return name.charAt(0).toUpperCase()
}
</script>

<style scoped>
.item-icon-wrapper {
  width: 42px; 
  height: 42px;
  background: rgba(0, 0, 0, 0.4); /* 深色半透明底座 */
  border: 1px solid #4c4d4f;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5px;
  box-sizing: border-box;
  flex-shrink: 0; /* 防止被挤压 */
}

.item-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  image-rendering: pixelated; /* 关键：像素风格 */
  filter: drop-shadow(2px 2px 2px rgba(0,0,0,0.6)); /* 立体阴影 */
}

.fallback-icon {
  color: #909399;
  font-weight: 800;
  font-size: 1.4em;
  text-transform: uppercase;
}
</style>
