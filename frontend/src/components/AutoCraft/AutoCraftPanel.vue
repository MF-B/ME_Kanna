<template>
  <div class="autocraft-panel">
    <!-- Grid Layout -->
    <div class="task-grid-container" v-if="sortedCards.length">
      <div
        v-for="(card, index) in sortedCards"
        :key="card.itemId"
        class="task-card"
        :class="{ dimmed: !card.hasTask || !card.isActive }"
        @click="handleCardClick(card)"
        :style="{ animationDelay: index * 0.03 + 's' }"
      >
        <!-- Main Card Content -->
        <div class="card-content">
          <ItemIcon :item-id="card.itemId" class="task-icon" />
          
          <!-- Status Triangle (Rotated Square) -->
          <div 
            v-if="shouldShowTriangle(card)"
            class="status-triangle" 
            :style="{ backgroundColor: getStatusColor(card) }"
          ></div>

          <!-- Overlay Quantity -->
          <div class="overlay-count">{{ formatCompact(getStock(card)) }}</div>
        </div>
      </div>
    </div>

    <div v-else class="brutalist-empty">
      <div class="empty-icon">⚙</div>
      <p>{{ craftablesLoading ? t('CRAFT.LOADING') || '加载中...' : t('CRAFT.NO_TASKS') }}</p>
    </div>

    <!-- Detail Dialog -->
    <el-dialog
      :model-value="detailVisible"
      @update:model-value="updateDetailVisible"
      width="320px" 
      class="brutalist-dialog small-dialog"
      :append-to-body="true"
    >
       <template #header>
         <div class="b-dialog-header centered-header">
           {{ detailTask ? displayItemName(detailTask.itemId) : t('CRAFT.TASK_DETAILS') }}
         </div>
       </template>
       
       <div v-if="detailTask" class="detail-content">
          <div class="b-summary vertical">
             <ItemIcon :item-id="detailTask.itemId" class="large-icon" />
             <div class="info-center">
                <p class="mono-id">{{ detailTask.itemId }}</p>
                <div class="stock-display">
                  {{ t('CRAFT.STOCK') }}: <span :style="{ color: getStockColor(detailTask) }">{{ formatCompact(getStock(detailTask)) }}</span>
                </div>
             </div>
          </div>
          
          <div class="b-inputs vertical">
             <div class="inp-row">
               <label>{{ t('CRAFT.MIN') }}</label>
               <el-input-number v-model="localForm.min" />
             </div>
             <div class="inp-row">
               <label>{{ t('CRAFT.MAX') }}</label>
               <el-input-number v-model="localForm.max" />
             </div>
          </div>

          <div class="dialog-actions">
             <div class="left-actions">
                <el-button v-if="detailTask.hasTask" type="danger" plain class="delete-btn" @click="handleDeleteTask(detailTask.itemId)">{{ t('CRAFT.DELETE') }}</el-button>
                <div class="b-switch-rect" :class="{ active: localForm.active }" @click="localForm.active = !localForm.active">
                   {{ localForm.active ? 'ON' : 'OFF' }}
                </div>
             </div>
             <el-button type="primary" :loading="detailSaving" @click="handleSaveLocal">{{ t('CRAFT.SAVE') }}</el-button>
          </div>
       </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useSystemStore } from '../../stores/systemStore'
import ItemIcon from '../ItemIcon.vue'
import { useI18n } from '../../composables/useI18n'

const systemStore = useSystemStore()
const { t } = useI18n()

const {
  allCraftableCards,
  inventoryIndex,
  itemRates,
  craftablesLoading,
  detailVisible,
  detailTask,
  detailMinThreshold,
  detailMaxThreshold,
  detailSaving
} = storeToRefs(systemStore)

const {
  formatCompact,
  displayItemName,
  openCardDetail,
  setDetailVisible,
  setDetailMinThreshold,
  setDetailMaxThreshold,
  saveTaskThresholds: _saveTaskThresholds,
  deleteTask: _deleteTask,
  handleTaskActiveChange: _handleTaskActiveChange
} = systemStore

// Get stock for a card: prefer inventoryIndex (live WebSocket), fallback to craftable count
function getStock(card) {
  const liveStock = inventoryIndex.value[card.itemId]
  if (liveStock !== undefined && liveStock !== null) return liveStock
  return card.craftableCount || 0
}

// Helper to determine status type for sorting and color
// Returns: 'YELLOW', 'RED', 'NONE', 'GRAY'
function getTaskStatusType(card) {
  // Unconfigured cards are always gray
  if (!card.hasTask) return 'GRAY'

  const stock = getStock(card)
  
  if (card.isActive) {
    if (stock < card.minThreshold) {
      const rate = itemRates.value[card.itemId] || 0
      if (rate > 0) return 'YELLOW' // Crafting
      return 'RED' // Waiting / No Rate
    }
    return 'NONE' // Active but Good Stock
  } else {
    return 'GRAY'
  }
}

function shouldShowTriangle(card) {
  const type = getTaskStatusType(card)
  return type !== 'NONE'
}

function getStatusColor(card) {
  const type = getTaskStatusType(card)
  switch (type) {
    case 'YELLOW': return '#e6a23c'
    case 'RED': return '#f56c6c'
    case 'GRAY': return '#909399'
    default: return 'transparent'
  }
}

// Sort cards based on status priority
const sortedCards = computed(() => {
  const cards = [...allCraftableCards.value]
  // Priority: Yellow(0) -> Red(1) -> None(2) -> Gray(3)
  const priority = { 'YELLOW': 0, 'RED': 1, 'NONE': 2, 'GRAY': 3 }
  
  return cards.sort((a, b) => {
    const typeA = getTaskStatusType(a)
    const typeB = getTaskStatusType(b)
    if (priority[typeA] !== priority[typeB]) {
      return priority[typeA] - priority[typeB]
    }
    // Secondary: tasks first, then non-tasks
    if (a.hasTask !== b.hasTask) return a.hasTask ? -1 : 1
    // Tertiary sort by itemId
    return (a.itemId || '').localeCompare(b.itemId || '')
  })
})

function getStockColor(task) {
   const stock = getStock(task)
   if (task.hasTask && stock < task.minThreshold) return 'var(--secondary-color)'
   if (task.hasTask && stock > task.maxThreshold) return 'var(--accent-color)'
   return 'var(--text-color)'
}

function handleCardClick(card) {
  openCardDetail(card)
}

async function handleDeleteTask(itemId) {
  try { await _deleteTask(itemId); ElMessage.success('DELETED'); setDetailVisible(false) } catch (err) { ElMessage.error(err.message) }
}

const localForm = ref({
  min: 1,
  max: 1,
  active: false
})

// Sync when dialog opens
watch(detailVisible, (newVal) => {
  if (newVal && detailTask.value) {
    localForm.value = {
      min: detailMinThreshold.value,
      max: detailMaxThreshold.value,
      active: detailTask.value.isActive
    }
  }
})

async function handleSaveLocal() {
  try {
    // 1. Update thresholds in store
    setDetailMinThreshold(localForm.value.min)
    setDetailMaxThreshold(localForm.value.max)
    
    // 2. Save thresholds (will also create task if new)
    await _saveTaskThresholds()
    
    // 3. Update active state if changed
    if (localForm.value.active !== detailTask.value.isActive) {
      await _handleTaskActiveChange(detailTask.value, localForm.value.active)
    }
    
    ElMessage.success(t('CRAFT.SAVE') + ' OK')
    setDetailVisible(false)
  } catch (err) {
    ElMessage.error(err.message)
  }
}

const updateDetailVisible = (v) => setDetailVisible(v)
</script>

<style scoped lang="scss">
.autocraft-panel {
  padding: 10px;
}

.task-grid-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  gap: 10px;
}

/* Responsive card sizing */
@media (min-width: 768px) {
  .task-grid-container {
    grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
    gap: 12px;
  }
}

@media (min-width: 1200px) {
  .task-grid-container {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 15px;
  }
}

.task-card {
  background: var(--surface-color);
  border: 2px solid var(--border-color);
  display: flex;
  flex-direction: column;
  cursor: pointer;
  transition: all 0.1s;
  box-shadow: 3px 3px 0 var(--border-color);
  aspect-ratio: 1;
  overflow: hidden;
  
  &:hover {
    transform: translate(1.5px, 1.5px);
    box-shadow: 1.5px 1.5px 0 var(--border-color);
  }
  
  &:active {
    transform: translate(3px, 3px);
    box-shadow: 0 0 0 var(--border-color);
  }

  &.dimmed {
    opacity: 0.55;
    
    &:hover {
      opacity: 0.8;
    }
  }
}

.card-content {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.task-icon {
  width: 60%;
  height: 60%;
}

.status-triangle {
  position: absolute;
  top: -16px;
  right: -16px;
  width: 32px;
  height: 32px;
  transform: rotate(45deg);
  border-bottom: 2px solid #000;
  box-shadow: -1px 1px 0 rgba(0,0,0,0.1);
  z-index: 1;
}

.overlay-count {
  position: absolute;
  bottom: 7px;
  right: 4px;
  font-family: var(--font-nums);
  font-size: 1.4rem;
  color: #ffffff;
  text-align: right;
  z-index: 2;
  pointer-events: none;
  line-height: 1;
  text-shadow: 
    -2.5px -2.5px 0 #5d4037,  
     2.5px -2.5px 0 #5d4037,
    -2.5px  2.5px 0 #5d4037,
     2.5px  2.5px 0 #5d4037,
    -2.5px 0 0 #5d4037,
     2.5px 0 0 #5d4037,
     0 -2.5px 0 #5d4037,
     0 2.5px 0 #5d4037;
}

@media (max-width: 767px) {
  .overlay-count {
    font-size: 1.2rem;
    text-shadow: 
      -2px -2px 0 #5d4037,  
       2px -2px 0 #5d4037,
      -2px  2px 0 #5d4037,
       2px  2px 0 #5d4037,
      -2px 0 0 #5d4037,
       2px 0 0 #5d4037,
       0 -2px 0 #5d4037,
       0 2px 0 #5d4037;
  }
  .status-triangle {
    top: -14px;
    right: -14px;
    width: 28px;
    height: 28px;
  }
}

/* Empty state */
.brutalist-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  color: #888;
  
  .empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }
  
  p {
    font-weight: bold;
    font-size: 1rem;
  }
}

/* Detail Dialog */
.b-summary.vertical {
  flex-direction: column;
  text-align: center;
  
  .large-icon { width: 64px; height: 64px; margin-bottom: 10px; }
  .info-center {
    h4 { margin: 5px 0; font-size: 1.2rem; }
    .mono-id { font-family: monospace; color: #888; font-size: 0.8rem; margin: 0; }
    .stock-display { margin-top: 10px; font-weight: bold; font-size: 1.1rem; }
  }
}

.b-inputs.vertical {
  flex-direction: column;
  gap: 10px;
  margin: 10px 0;
  
  .inp-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    label { flex: 1; font-weight: bold; }
  }
}

.dialog-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  border-top: 2px dashed var(--secondary-color);
  padding-top: 10px;
  
  .left-actions {
    display: flex;
    gap: 15px;
    align-items: center;
  }
}

.b-switch-rect {
  border: 2px solid var(--border-color);
  background: var(--surface-color);
  padding: 5px 10px;
  font-weight: bold;
  cursor: pointer;
  box-shadow: 2px 2px 0 var(--border-color);
  transition: all 0.1s;
  
  &.active {
    background: var(--accent-color);
    color: #fff;
  }
  
  &:hover { transform: translate(1px, 1px); box-shadow: 1px 1px 0 var(--border-color); }
  &:active { transform: translate(2px, 2px); box-shadow: 0 0 0 var(--border-color); }
}

.delete-btn {
  background-color: var(--danger-color, #f56c6c) !important;
  color: white !important;
  font-weight: bold;
  border: 2px solid var(--border-color) !important;
}

.b-summary { display: flex; gap: 15px; align-items: center; background: var(--bg-color); border: 2px solid var(--border-color); padding: 10px; margin-bottom: 10px; }

.centered-header {
  text-align: center;
  border-bottom: 2px dashed #e6a23c !important;
}

/* Equal height and consistent style for action buttons */
.dialog-actions {
  .el-button, .b-switch-rect {
    height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    vertical-align: middle;
    border: 2px solid var(--border-color) !important;
    font-weight: bold;
  }
  
  .b-switch-rect {
    margin-left: 0;
    padding: 0 15px;
    min-width: 55px;
    box-shadow: 4px 4px 0 var(--border-color);
    
    &:hover {
      transform: translate(2px, 2px);
      box-shadow: 2px 2px 0 var(--border-color);
    }
    &:active {
      transform: translate(4px, 4px);
      box-shadow: 0 0 0 var(--border-color);
    }
  }
  
  .el-button--primary {
    border: 2px solid var(--border-color) !important;
  }
}
</style>

<style lang="scss">
/* Global styles for appended dialogs */
.brutalist-dialog .el-dialog__headerbtn {
  top: 5px !important;
  right: 5px !important;
  width: 30px;
  height: 30px;
  background: transparent; 
  border: none;
  display: flex !important;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  
  .el-dialog__close {
    color: var(--text-color);
    font-weight: 900 !important;
    font-size: 1.2rem !important;
  }
  
  &:hover {
    background: var(--danger-color, #f56c6c);
    .el-dialog__close { color: white !important; }
  }
}

/* Ensure symmetric padding for header and body so dashed lines align */
.small-dialog .el-dialog__header {
  padding: 20px !important;
  margin: 0 !important;
}

.small-dialog .el-dialog__body {
  padding: 20px !important;
}

/* Hide close button for detail dialog */
.small-dialog .el-dialog__headerbtn {
  display: none !important;
}

/* Constrain dialog to viewport */
.brutalist-dialog.el-dialog {
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  max-width: 95vw;
}

.brutalist-dialog.el-dialog > .el-dialog__body {
  overflow-y: auto;
  flex: 1;
  min-height: 0;
}
</style>
