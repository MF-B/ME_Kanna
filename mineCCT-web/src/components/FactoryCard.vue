<template>
  <div class="factory-accordion" :class="{ expanded: isExpanded }">
    <!-- Header / Collapsed State -->
    <div class="accordion-header" @click="toggleExpand">
      <div class="header-top">
        <h3 class="factory-name">{{ factory.name || t('FACTORY.UNIT') }}</h3>
      </div>

      <div class="header-bottom" @click.stop>
        <div class="icon-btn expand-btn" @click="toggleExpand">
           <span class="arrow-char" :class="{ rotated: isExpanded }">›</span>
        </div>
        
        <div class="icon-btn settings-btn" @click.stop="openSettings">
           ⚙
        </div>

        <div class="b-toggle-btn" :class="{ active: localActive }" @click="toggleSwitch">
           <span class="toggle-text">{{ localActive ? 'ON' : 'OFF' }}</span>
        </div>
      </div>
    </div>

    <!-- Expanded Body -->
    <div v-show="isExpanded" class="accordion-body">
      <div class="item-grid-container" v-if="visibleItems.length">
        <div class="item-grid">
           <div 
             v-for="item in visibleItems" 
             :key="item.itemId" 
             class="grid-cell" 
             :title="getItemName(item.itemId)"
           >
              <ItemIcon :item-id="item.itemId" class="cell-icon" />
              <div class="overlay-rate">{{ formatNumber(item.prodRate) }}</div>
              <div class="overlay-count">{{ formatCompact(item.count) }}</div>
           </div>
        </div>
      </div>
      <div v-else class="empty-state">{{ t('FACTORY.NO_OUTPUT') }}</div>
    </div>
  </div>

  <!-- Settings Dialog -->
  <el-dialog
    v-model="dialogVisible"
    :title="t('FACTORY.CONFIG')"
    width="520px"
    class="brutalist-dialog"
    :append-to-body="true"
  >
    <div class="b-form-group">
      <label>ID</label>
      <div class="static-value">{{ factory.id }}</div>
    </div>
    <div class="b-form-group">
      <label>{{ t('FACTORY.UNIT_NAME') }}</label>
      <el-input
        v-model="localName"
        :placeholder="t('FACTORY.ENTER_NAME')"
        @input="nameEditing = true"
        @change="commitName"
        @blur="commitName"
      />
    </div>


  </el-dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import ItemIcon from './ItemIcon.vue'
import { useItemNames } from '../composables/useItemNames'
import { useI18n } from '../composables/useI18n'

const props = defineProps(['factory'])
const emit = defineEmits(['command'])
const { t } = useI18n()

const isExpanded = ref(false)
const loading = ref(false)
const localActive = ref(props.factory.isActive)

const dialogVisible = ref(false)

const localName = ref(props.factory?.name || '')
const nameEditing = ref(false)
const { names: itemNames } = useItemNames()

watch(() => props.factory.isActive, (newVal) => {
  localActive.value = newVal
  loading.value = false
})

watch(() => props.factory, (factory) => {
  if (!nameEditing.value && !dialogVisible.value) {
    localName.value = factory?.name || ''
  }
}, { deep: true, immediate: true })

function toggleExpand() {
  isExpanded.value = !isExpanded.value
}

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

function formatNumber(num) {
  if (!num) return '0'
  return formatCompact(num)
}

function formatCompact(num) {
  if (!num) return '0'
  if (num < 1000) return num.toLocaleString()
  return Intl.NumberFormat('en-US', {
    notation: "compact", 
    maximumFractionDigits: 1
  }).format(num)
}
</script>

<style scoped lang="scss">

.factory-accordion {
  border: 2px solid var(--border-color);
  background: var(--bg-color);
  box-shadow: 4px 4px 0 var(--border-color);
  transition: all 0.1s;
  
  /* Hover: Move towards shadow (press effect) */
  &:hover {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 var(--border-color);
  }

  /* Expanded: Keep the pressed effect */
  &.expanded {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 var(--border-color);
  }
}

/* Header style */
.accordion-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 15px;
  cursor: pointer;
  background: var(--surface-color);
  border-bottom: 2px solid transparent; 
  transition: all 0.1s;
  /* Removed click press effect on header itself to avoid double-movement */
}

.factory-accordion.expanded .accordion-header {
  border-bottom: 2px dashed var(--border-color);
}


.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.factory-name {
  margin: 0;
  font-size: 1.2rem;
  font-weight: bold;
}

.factory-id {
  font-family: monospace;
  font-size: 0.8rem;
  color: #888;
}

.header-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between; /* Arrow Left, Settings Center, Switch Right */
  width: 100%;
}

.icon-btn {
  font-size: 1.2rem;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s;
  
  &:hover { transform: scale(1.1); color: var(--primary-color); }
}

.arrow-char {
  display: inline-block;
  font-size: 1.5rem;
  line-height: 1;
  font-weight: bold;
  transition: transform 0.3s;
}

.arrow-char.rotated {
  transform: rotate(90deg);
}

.b-toggle-btn {
  width: 50px;
  height: 26px;
  background: var(--surface-color);
  border: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 2px 2px 0 var(--border-color);
  transition: all 0.1s;
  
  .toggle-text {
    font-size: 0.8rem;
    font-weight: 900;
    color: var(--text-color);
  }
  
  &:hover {
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0 var(--border-color);
  }
  
  &:active {
    transform: translate(2px, 2px);
    box-shadow: none;
  }
  
  &.active {
    background: var(--accent-color);
    .toggle-text { color: #fff; }
  }
}

.accordion-body {
  padding: 8px;
  background: var(--bg-color);
}

.item-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 4px;
  padding: 2px;
  background: #cbccd4;
  border: 2px solid #f2f2f2;
}

.grid-cell {
  position: relative;
  aspect-ratio: 1;
  background: #adb0c4;
  border: none;
  box-shadow: inset 0 2px 0 #9a9fb4;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  transition: background 0.05s;
  
  &:hover {
    background: #a6bedc;
    outline: 2px solid #daffff;
    z-index: 1;
  }
}

.cell-icon {
  width: 60%;
  height: 60%;
  opacity: 1.0; /* Restored full opacity */
  /* filter: grayscale(100%); REMOVED */
}

/* Nanotype Font for numbers */
.overlay-rate {
  position: absolute;
  top: -5px;
  left: 3px;
  font-family: var(--font-nums);
  font-size: 1rem;
  color: #ffffff;
  z-index: 2;
  pointer-events: none;
  line-height: 1;
  font-weight: normal;
  text-shadow: 
    -1.5px -1.5px 0 #2e7d32,  
     1.5px -1.5px 0 #2e7d32,
    -1.5px  1.5px 0 #2e7d32,
     1.5px  1.5px 0 #2e7d32,
    -1.5px 0 0 #2e7d32,
     1.5px 0 0 #2e7d32,
     0 -1.5px 0 #2e7d32,
     0 1.5px 0 #2e7d32;
}

.overlay-count {
  position: absolute;
  bottom: 5px;
  right: 1px;
  font-family: var(--font-nums);
  font-size: 1rem;
  color: #ffffff;
  text-align: right;
  z-index: 2;
  pointer-events: none;
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
}

.empty-state, .empty-text {
  color: #666;
  text-align: center;
  padding: 10px;
}

/* Dialog */
.b-form-group {
  margin-bottom: 15px;
  label { display: block; font-weight: bold; margin-bottom: 5px; }
  .static-value {
    padding: 5px 10px;
    background: var(--bg-color);
    border: 2px solid var(--border-color);
    font-family: monospace;
    color: #888;
  }
}
.b-list {
  border: 2px solid var(--border-color);
  max-height: 300px;
  overflow-y: auto;
}
.b-list-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  border-bottom: 1px solid var(--border-color);
}
</style>
