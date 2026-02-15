<template>
  <div class="tree-node">
    <div class="node-row" :class="statusClass">
      <ItemIcon :item-id="node.itemId" />
      <div class="node-meta">
        <div class="node-title">
          <span>{{ node.itemName || node.itemId }}</span>
          <span v-if="node.count && node.count > 1" class="node-count">x{{ node.count }}</span>
        </div>
        <div class="node-sub">
          <span>CUR: {{ formatCompact(currentCount) }}</span>
          <span v-if="threshold" class="node-target">TARGET: {{ threshold.min }} / {{ threshold.max }}</span>
          <span v-else class="node-target muted">NO TARGET</span>
        </div>
      </div>
      <div class="node-badge">
        {{ statusLabel }}
      </div>
    </div>

    <div v-if="node.children && node.children.length" class="node-children">
      <AutoCraftTree
        v-for="child in node.children"
        :key="child.itemId + ':' + (child.count || 1)"
        :node="child"
        :inventory-index="inventoryIndex"
        :task-index="taskIndex"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import ItemIcon from './ItemIcon.vue'

defineOptions({ name: 'AutoCraftTree' })

const props = defineProps({
  node: { type: Object, required: true },
  inventoryIndex: { type: Object, required: true },
  taskIndex: { type: Object, required: true }
})

const currentCount = computed(() => {
  return props.inventoryIndex[props.node.itemId] || 0
})

const threshold = computed(() => {
  const matchedTask = props.taskIndex[props.node.itemId]
  if (!matchedTask) return null
  return { min: matchedTask.minThreshold, max: matchedTask.maxThreshold }
})

const statusLabel = computed(() => {
  if (!threshold.value) return 'NO TARGET'
  if (currentCount.value < threshold.value.min) return 'LOW'
  if (currentCount.value >= threshold.value.max) return 'OK'
  return 'MID'
})

const statusClass = computed(() => {
  if (!threshold.value) return 'status-neutral'
  if (currentCount.value < threshold.value.min) return 'status-low'
  if (currentCount.value >= threshold.value.max) return 'status-ok'
  return 'status-mid'
})

function formatCompact(value) {
  if (value === null || value === undefined) return '0'
  if (value < 10000) return value.toLocaleString()
  return Intl.NumberFormat('en-US', {
    notation: 'compact',
    maximumFractionDigits: 1
  }).format(value)
}
</script>

<style scoped lang="scss">
.tree-node {
  position: relative;
  padding-left: 20px;
}

.node-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 12px;
  align-items: center;
  background: #000;
  border: 2px solid #333;
  padding: 8px 12px;
  margin-bottom: 10px;
  transition: all 0.2s;
  
  &:hover {
    border-color: #fff;
    transform: translateX(5px);
  }
}

.node-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.node-title {
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 0.9rem;
  
  span {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.node-count {
  color: var(--primary-color, #FFD600);
  font-size: 0.8rem;
}

.node-sub {
  display: flex;
  gap: 12px;
  font-size: 0.75rem;
  color: #888;
  font-family: monospace;
}

.node-badge {
  font-size: 0.7rem;
  font-weight: 900;
  padding: 2px 6px;
  background: #fff;
  color: #000;
  border: 2px solid #000;
}

/* Status variants */
.status-low {
  border-left: 5px solid var(--secondary-color, #FF5722);
  
  .node-badge {
    background: var(--secondary-color, #FF5722);
    color: #fff;
  }
}

.status-ok {
  border-left: 5px solid var(--accent-color, #00E676);
  
  .node-badge {
    background: var(--accent-color, #00E676);
    color: #000;
  }
}

.status-mid {
  border-left: 5px solid var(--primary-color, #FFD600);
}

.status-neutral {
  border-left: 2px solid #333;
}

.node-children {
  margin-left: 20px;
  border-left: 2px dashed #333;
  padding-left: 10px;
}
</style>
