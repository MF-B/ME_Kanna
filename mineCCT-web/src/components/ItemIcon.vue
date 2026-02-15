<template>
  <div class="item-icon-container">
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
    <!-- 右下角数量显示 -->
    <div v-if="count !== undefined && count !== null" class="item-count">
      {{ formatCount(count) }}
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  itemId: { type: String, required: false },
  count: { type: [Number, String], required: false }
})

const isError = ref(false)

watch(() => props.itemId, () => {
  isError.value = false
})

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

// 格式化数量，参考 MC 风格 (如 1000 -> 1k)
const formatCount = (val) => {
  const num = Number(val)
  if (isNaN(num)) return val
  if (num < 1000) return num.toString()
  if (num < 1000000) return (num / 1000).toFixed(1).replace(/\.0$/, '') + 'k'
  return (num / 1000000).toFixed(1).replace(/\.0$/, '') + 'M'
}
</script>

<style scoped>
.item-icon-container {
  width: 32px; 
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  image-rendering: pixelated;
}

.item-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.item-count {
  position: absolute;
  bottom: -1px;
  right: -4px;
  color: #ffffff;
  font-family: var(--font-nums);
  font-size: 1rem;
  line-height: 1;
  text-shadow: 
    -1.5px -1.5px 0 #5d4037,  
     1.5px -1.5px 0 #5d4037,
    -1.5px  1.5px 0 #5d4037,
     1.5px  1.5px 0 #5d4037,
    -1.5px 0 0 #5d4037,
     1.5px 0 0 #5d4037,
     0 -1.5px 0 #5d4037,
     0 1.5px 0 #5d4037;
  pointer-events: none;
  z-index: 2;
}

.fallback-icon {
  color: #ffffff;
  font-weight: bold;
  font-size: 1.2em;
  text-shadow: 1px 1px 0px #3f3f3f;
}
</style>
