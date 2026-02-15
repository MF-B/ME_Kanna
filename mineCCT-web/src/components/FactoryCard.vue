<template>
  <BrutalistCard :title="factory.name || 'FACTORY.UNIT'" class="factory-card-wrapper">
    <template #header>
      <div class="brutalist-header">
        <h3 class="card-title">{{ factory.name || 'FACTORY' }}</h3>
        <div class="status-indicator" :class="{ running: factory.isActive }">
          {{ factory.isActive ? 'RUN' : 'STOP' }}
        </div>
      </div>
    </template>

    <div class="factory-details">
      <div class="id-row">ID: {{ factory.id }}</div>
      
      <div class="item-list-container" v-if="visibleItems.length">
        <div class="b-item-row" v-for="item in visibleItems" :key="item.itemId">
          <ItemIcon :item-id="item.itemId" class="b-icon" />
          <div class="b-meta">
            <div class="b-name">{{ getItemName(item.itemId) }}</div>
            <div class="b-stats">
              <span class="rate">{{ formatNumber(item.prodRate) }}/h</span>
              <span class="count">{{ formatCompact(item.count) }}</span>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">NO OUTPUT</div>
    </div>

    <div class="factory-controls">
      <div class="control-row" @click.stop>
        <span class="lbl">POWER</span>
        <div class="b-switch" :class="{ active: localActive }" @click="toggleSwitch">
          <div class="track"></div>
          <div class="knob"></div>
        </div>
      </div>
      <div class="control-row settings-btn" @click.stop="openSettings">
        <span class="lbl">CONFIG</span>
        <span class="icon">⚙</span>
      </div>
    </div>
  </BrutalistCard>

  <el-dialog
    v-model="dialogVisible"
    title="CONFIGURATION"
    width="520px"
    class="brutalist-dialog"
    :append-to-body="true"
  >
    <div class="b-form-group">
      <label>UNIT NAME</label>
      <el-input
        v-model="localName"
        placeholder="ENTER NAME"
        @input="nameEditing = true"
        @change="commitName"
        @blur="commitName"
      />
    </div>

    <div class="b-form-group">
      <label>OUTPUTS</label>
      <div v-if="localItems.length" class="b-list">
        <div
          class="b-list-item"
          v-for="(item, index) in localItems"
          :key="item.itemId"
          draggable="true"
          @dragstart="handleDragStart(index)"
          @dragover.prevent="handleDragOver(index)"
          @drop="handleDrop"
        >
          <div class="drag-handle">::</div>
          <el-checkbox v-model="item.visible" @change="commitSettings"></el-checkbox>
          <span class="item-name">{{ getItemName(item.itemId) }}</span>
        </div>
      </div>
      <div v-else class="empty-text">NO DATA</div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import BrutalistCard from './BrutalistCard.vue'
import ItemIcon from './ItemIcon.vue'
import { useItemNames } from '../composables/useItemNames'

const props = defineProps(['factory'])
const emit = defineEmits(['command'])

const loading = ref(false)
const localActive = ref(props.factory.isActive)
const localItems = ref([])
const dialogVisible = ref(false)
const dragIndex = ref(null)
const localName = ref(props.factory?.name || '')
const nameEditing = ref(false)
const { names: itemNames } = useItemNames()

watch(() => props.factory.isActive, (newVal) => {
  localActive.value = newVal
  loading.value = false
})

watch(() => props.factory, (factory) => {
  const items = normalizeItems(factory?.items)
  localItems.value = sortItems(items).map((item) => ({ ...item }))
  if (!nameEditing.value && !dialogVisible.value) {
    localName.value = factory?.name || ''
  }
}, { deep: true, immediate: true })

function toggleSwitch() {
  if (loading.value) return
  const newVal = !localActive.value
  localActive.value = newVal
  handleSwitchChange(newVal)
}

function handleSwitchChange(val) {
  loading.value = true
  const action = val ? 'start' : 'stop'
  emit('command', { 
    target: props.factory.id, 
    action: action 
  })
  setTimeout(() => { loading.value = false }, 2000)
}

function openSettings() {
  dialogVisible.value = true
  localName.value = props.factory?.name || ''
  nameEditing.value = false
}

function handleDragStart(index) {
  dragIndex.value = index
}

function handleDragOver(index) {
  if (dragIndex.value === null || dragIndex.value === index) return
  const updated = [...localItems.value]
  const [moved] = updated.splice(dragIndex.value, 1)
  updated.splice(index, 0, moved)
  localItems.value = updated
  dragIndex.value = index
}

function handleDrop() {
  dragIndex.value = null
  commitSettings()
}

function commitName() {
  const nextName = localName.value?.trim()
  if (!nextName) return
  emit('command', {
    target: props.factory.id,
    action: 'update_factory_name',
    name: nextName
  })
  nameEditing.value = false
}

const itemsList = computed(() => normalizeItems(props.factory?.items))
const sortedItems = computed(() => sortItems(itemsList.value))

const visibleItems = computed(() => {
  return sortedItems.value.filter((factoryItem) => factoryItem.visible)
})

function normalizeItems(rawItems) {
  if (!rawItems) return []
  const itemList = Array.isArray(rawItems) ? rawItems : Object.values(rawItems)
  return itemList.map((factoryItem) => ({
    ...factoryItem,
    visible: factoryItem.visible !== false,
    order: Number.isFinite(factoryItem.order) ? factoryItem.order : 0
  }))
}

function sortItems(itemList) {
  const hasOrder = itemList.some((factoryItem) => factoryItem.order && factoryItem.order > 0)
  const byItemId = (leftItem, rightItem) => (leftItem.itemId || '').localeCompare(rightItem.itemId || '')
  if (!hasOrder) {
    return [...itemList].sort(byItemId)
  }
  return [...itemList].sort((leftItem, rightItem) => {
    const leftOrder = leftItem.order && leftItem.order > 0 ? leftItem.order : Number.POSITIVE_INFINITY
    const rightOrder = rightItem.order && rightItem.order > 0 ? rightItem.order : Number.POSITIVE_INFINITY
    if (leftOrder !== rightOrder) return leftOrder - rightOrder
    return byItemId(leftItem, rightItem)
  })
}

function getItemName(itemId) {
  return itemNames[itemId] || itemId
}

function commitSettings() {
  emit('command', {
    target: props.factory.id,
    action: 'update_factory_items',
    items: localItems.value.map((factoryItem, index) => ({
      itemId: factoryItem.itemId,
      visible: factoryItem.visible,
      order: index + 1
    }))
  })
}

function formatNumber(num) {
  if (!num) return '0'
  return Math.floor(num).toLocaleString()
}

function formatCompact(num) {
  if (!num) return '0'
  if (num < 10000) return num.toLocaleString()
  return Intl.NumberFormat('en-US', {
    notation: "compact",
    maximumFractionDigits: 1
  }).format(num)
}
</script>

<style scoped lang="scss">
.brutalist-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #fff;
  padding-bottom: 10px;
  margin-bottom: 10px;
  
  .card-title {
    margin: 0;
    font-size: 1rem;
    background: var(--primary-color);
    color: #000;
    padding: 2px 5px;
  }
}

.status-indicator {
  font-weight: bold;
  font-size: 0.8rem;
  padding: 2px 6px;
  background: #333;
  color: #fff;
  border: 1px solid #fff;
  
  &.running {
    background: var(--accent-color);
    color: #000;
  }
}

.factory-details {
  .id-row {
    font-size: 0.7rem;
    color: #888;
    margin-bottom: 10px;
    font-family: monospace;
  }
}

.item-list-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 15px;
  max-height: 200px;
  overflow-y: auto;
}

.b-item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #000;
  border: 1px solid #333;
  padding: 5px;
  
  &:hover {
    border-color: var(--primary-color);
  }
}

.b-meta {
  flex: 1;
  overflow: hidden;
  
  .b-name {
    font-size: 0.9rem;
    font-weight: bold;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .b-stats {
    font-size: 0.75rem;
    color: #888;
    display: flex;
    justify-content: space-between;
  }
}

.factory-controls {
  border-top: 2px solid #fff;
  padding-top: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.control-row {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  
  .lbl {
    font-size: 0.8rem;
    font-weight: bold;
  }
  
  &.settings-btn:hover {
    color: var(--primary-color);
  }
}

.b-switch {
  width: 40px;
  height: 20px;
  background: #333;
  border: 2px solid #fff;
  position: relative;
  cursor: pointer;
  
  .knob {
    width: 12px;
    height: 12px;
    background: #fff;
    position: absolute;
    top: 2px;
    left: 2px;
    transition: all 0.2s;
  }
  
  &.active {
    background: var(--accent-color);
    
    .knob {
      left: 20px;
      background: #000;
    }
  }
}

.empty-state, .empty-text {
  color: #666;
  font-style: italic;
  text-align: center;
  padding: 10px;
}

/* Dialog Styles */
.b-form-group {
  margin-bottom: 15px;
  
  label {
    display: block;
    font-size: 0.8rem;
    font-weight: bold;
    margin-bottom: 5px;
    color: var(--primary-color);
  }
}

.b-list {
  border: 2px solid #333;
  max-height: 300px;
  overflow-y: auto;
}

.b-list-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  border-bottom: 1px solid #333;
  background: #000;
  
  &:last-child { border-bottom: none; }
  
  .drag-handle {
    cursor: grab;
    color: #666;
    font-weight: bold;
  }
  
  .item-name {
    font-size: 0.9rem;
  }
}
</style>
