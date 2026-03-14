<template>
  <div class="tree-node">
    <div class="node-row" :class="statusClass">
      <ItemIcon :item-id="node.itemId" />
      <div class="node-meta">
        <div class="node-title">
          <span>{{ getItemName(node.itemId) }}</span>
          <span v-if="node.count && node.count > 1" class="node-count">x{{ node.count }}</span>
        </div>
        <div class="node-sub">
          <span>{{ t('CRAFT.CUR') }}: {{ formatCompact(currentCount) }}</span>
          <span v-if="threshold" class="node-target">{{ t('CRAFT.TARGET') }}: {{ threshold.min }} / {{ threshold.max }}</span>
          <span v-else class="node-target muted">{{ t('TREE.NO_TARGET') }}</span>
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
import { useI18n } from '../composables/useI18n'
import { useItemInfo } from '../composables/useItemInfo'

defineOptions({ name: 'AutoCraftTree' })

const props = defineProps({
  node: { type: Object, required: true },
  inventoryIndex: { type: Object, required: true },
  taskIndex: { type: Object, required: true }
})

const { t } = useI18n()
const { names: itemNames, ensureName } = useItemInfo()

function getItemName(itemId) {
  if (!itemId) return ''
  ensureName(itemId)
  return itemNames[itemId] || itemId
}

const currentCount = computed(() => {
  return props.inventoryIndex[props.node.itemId] || 0
})

const threshold = computed(() => {
  const matchedTask = props.taskIndex[props.node.itemId]
  if (!matchedTask) return null
  return { min: matchedTask.minThreshold, max: matchedTask.maxThreshold }
})

const statusLabel = computed(() => {
  if (!threshold.value) return t('TREE.NO_TARGET')
  if (currentCount.value < threshold.value.min) return t('TREE.LOW')
  if (currentCount.value >= threshold.value.max) return t('TREE.OK')
  return t('TREE.MID')
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
  background: var(--bg-color);
  border: 2px solid var(--border-color);
  padding: 8px 12px;
  margin-bottom: 10px;
  transition: all 0.2s;
  color: var(--text-color);
  
  &:hover {
    border-color: var(--primary-color);
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
  color: var(--text-color);
  font-size: 0.9rem;
  
  span {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.node-count {
  color: var(--primary-color);
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
  background: var(--surface-color);
  color: var(--text-color);
  border: 2px solid var(--border-color);
}

/* Status variants */
.status-low {
  border-left: 5px solid var(--secondary-color);
  
  .node-badge {
    background: var(--secondary-color);
    color: #fff;
  }
}

.status-ok {
  border-left: 5px solid var(--accent-color);
  
  .node-badge {
    background: var(--accent-color);
    color: #000;
  }
}

.status-mid {
  border-left: 5px solid var(--primary-color);
}

.status-neutral {
  border-left: 2px solid var(--border-color);
}

.node-children {
  margin-left: 20px;
  border-left: 2px dashed var(--border-color);
  padding-left: 10px;
}
</style>
