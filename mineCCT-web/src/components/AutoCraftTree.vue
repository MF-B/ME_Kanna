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
          <span>现有 {{ formatCompact(currentCount) }}</span>
          <span v-if="threshold" class="node-target">目标 {{ threshold.min }} / {{ threshold.max }}</span>
          <span v-else class="node-target muted">未设置目标</span>
        </div>
      </div>
      <div class="node-badge" :class="statusClass">
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

<style scoped>
.tree-node {
  position: relative;
  padding-left: 18px;
}

.node-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 12px;
  align-items: center;
  background: rgba(16, 19, 28, 0.7);
  border: 1px solid rgba(120, 130, 170, 0.2);
  border-radius: 12px;
  padding: 10px 14px;
  margin-bottom: 10px;
}

.node-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.node-title {
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #eef2ff;
}

.node-title span {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-count {
  color: #9ae6b4;
  font-size: 0.85rem;
}

.node-sub {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
  color: #a0aec0;
}

.node-target.muted {
  color: #6b7280;
}

.node-badge {
  font-size: 0.7rem;
  letter-spacing: 1px;
  padding: 4px 8px;
  border-radius: 999px;
  text-transform: uppercase;
  border: 1px solid transparent;
}

.status-low .node-badge {
  background: rgba(255, 107, 107, 0.2);
  border-color: rgba(255, 107, 107, 0.5);
  color: #ff6b6b;
}

.status-mid .node-badge {
  background: rgba(255, 193, 7, 0.18);
  border-color: rgba(255, 193, 7, 0.4);
  color: #f8c04b;
}

.status-ok .node-badge {
  background: rgba(61, 214, 165, 0.2);
  border-color: rgba(61, 214, 165, 0.5);
  color: #3dd6a5;
}

.status-neutral .node-badge {
  background: rgba(148, 163, 184, 0.15);
  border-color: rgba(148, 163, 184, 0.35);
  color: #94a3b8;
}

.node-children {
  margin-left: 18px;
  border-left: 1px dashed rgba(148, 163, 184, 0.4);
  padding-left: 16px;
}
</style>
