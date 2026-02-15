<template>
  <div class="autocraft-panel">
    <!-- Grid Layout -->
    <div class="task-grid-container" v-if="sortedAutoCraftTasks.length">
      <div
        v-for="(task, index) in sortedAutoCraftTasks"
        :key="task.itemId"
        class="task-card"
        :class="{ pulse: !task.isActive }"
        @click="openTaskDetail(task)"
        :style="{ animationDelay: index * 0.05 + 's' }"
      >
        <!-- Main Card Content -->
        <div class="card-content">
          <ItemIcon :item-id="task.itemId" class="task-icon" />
          
          <!-- Status Triangle (Rotated Square) -->
          <div 
            v-if="shouldShowTriangle(task)"
            class="status-triangle" 
            :style="{ backgroundColor: getStatusColor(task) }"
          ></div>

          <!-- Overlay Quantity -->
          <div class="overlay-count">{{ formatCompact(inventoryIndex[task.itemId] || 0) }}</div>
        </div>
      </div>
    </div>

    <el-empty
      v-else
      :description="t('CRAFT.NO_TASKS')"
      class="brutalist-empty"
    />

    <!-- FAB -->
    <button class="brutalist-fab" @click="openWizard">
      +
    </button>

    <!-- Wizard Dialog (Unchanged logic, just simplified template if needed) -->
    <el-dialog
      :model-value="wizardVisible"
      @update:model-value="updateWizardVisible"
      width="640px"
      class="brutalist-dialog"
      :append-to-body="true"
    >
      <template #header>
        <div class="b-dialog-header">{{ t('CRAFT.NEW_TASK') }}</div>
      </template>

      <div v-if="wizardStep === 1">
        <div class="b-toolbar">
          <el-button @click="handleFetchCraftables">{{ t('CRAFT.REFRESH') }}</el-button>
          <el-input :model-value="craftableQuery" @update:model-value="updateCraftableQuery" :placeholder="t('CRAFT.SEARCH_ITEM')" clearable />
        </div>
        
        <div class="bg-grid">
          <div
            v-for="item in filteredCraftables"
            :key="item.itemId"
            class="grid-cell"
            :class="{ selected: selectedCraftable && selectedCraftable.itemId === item.itemId }"
            @click="selectCraftable(item)"
          >
            <ItemIcon 
              :item-id="item.itemId" 
              :count="inventoryIndex[item.itemId] !== undefined ? inventoryIndex[item.itemId] : (item.count || 0)" 
            />
          </div>
        </div>
      </div>

      <div v-else-if="wizardStep === 2">
        <div class="b-summary">
          <ItemIcon :item-id="selectedCraftable?.itemId" />
          <div class="info">
             <h4>{{ displayItemName(selectedCraftable?.itemId, selectedCraftable?.itemName) }}</h4>
             <p>{{ selectedCraftable?.itemId }}</p>
          </div>
        </div>
        
        <div class="b-inputs">
           <div class="inp-grp">
             <label>{{ t('CRAFT.MIN_THRESHOLD') }}</label>
             <el-input-number :model-value="minThreshold" @update:model-value="updateMinThreshold" :min="1" />
           </div>
           <div class="inp-grp">
             <label>{{ t('CRAFT.MAX_THRESHOLD') }}</label>
             <el-input-number :model-value="maxThreshold" @update:model-value="updateMaxThreshold" :min="1" />
           </div>
        </div>
      </div>

      <template #footer>
        <div class="b-footer">
          <el-button @click="closeWizard">{{ t('CRAFT.CANCEL') }}</el-button>
          <div class="actions">
            <el-button v-if="wizardStep > 1" @click="prevWizardStep">{{ t('CRAFT.BACK') }}</el-button>
            <el-button
              v-if="wizardStep < 2"
              type="primary"
              :disabled="wizardStep === 1 && !selectedCraftable"
              @click="nextWizardStep"
            >{{ t('CRAFT.NEXT') }}</el-button>
            <el-button
              v-if="wizardStep === 2"
              type="primary"
              :loading="recipeLoading"
              @click="handleFinishWizard"
            >{{ t('CRAFT.CREATE') }}</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- Simplified Detail Dialog -->
    <el-dialog
      :model-value="detailVisible"
      @update:model-value="updateDetailVisible"
      width="320px" 
      class="brutalist-dialog small-dialog"
      :append-to-body="true"
    >
       <template #header>
         <div class="b-dialog-header centered-header">
           {{ detailTask ? displayItemName(detailTask.itemId, detailTask.itemName) : t('CRAFT.TASK_DETAILS') }}
         </div>
       </template>
       
       <div v-if="detailTask" class="detail-content">
          <div class="b-summary vertical">
             <ItemIcon :item-id="detailTask.itemId" class="large-icon" />
             <div class="info-center">
                <p class="mono-id">{{ detailTask.itemId }}</p>
                <div class="stock-display">
                  {{ t('CRAFT.STOCK') }}: <span :style="{ color: getStockColor(detailTask) }">{{ formatCompact(inventoryIndex[detailTask.itemId] || 0) }}</span>
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
                <el-button type="danger" plain class="delete-btn" @click="handleDeleteTask(detailTask.itemId)">{{ t('CRAFT.DELETE') }}</el-button>
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
  autoCraftTasks,
  inventoryIndex,
  itemRates,
  taskIndex,
  wizardVisible,
  wizardStep,
  craftablesLoading,
  craftableQuery,
  selectedCraftable,
  filteredCraftables,
  minThreshold,
  maxThreshold,
  recipeLoading,
  detailVisible,
  detailTask,
  detailMinThreshold,
  detailMaxThreshold,
  detailSaving
} = storeToRefs(systemStore)

const {
  formatCompact,
  displayItemName,
  openWizard,
  closeWizard,
  selectCraftable,
  prevWizardStep,
  nextWizardStep,
  openTaskDetail,
  setWizardVisible,
  setCraftableQuery,
  setMinThreshold,
  setMaxThreshold,
  setDetailVisible,
  setDetailMinThreshold,
  setDetailMaxThreshold,
  saveTaskThresholds: _saveTaskThresholds,
  deleteTask: _deleteTask,
  handleTaskActiveChange: _handleTaskActiveChange,
  fetchCraftables: _fetchCraftables,
  finishWizard: _finishWizard
} = systemStore

// Helper to determine status type for sorting and color
// Returns: 'YELLOW', 'RED', 'NONE', 'GRAY'
function getTaskStatusType(task) {
  const stock = inventoryIndex.value[task.itemId] || 0
  
  if (task.isActive) {
    if (stock < task.minThreshold) {
      // Active & Low Stock
      const rate = itemRates.value[task.itemId] || 0
      if (rate > 0) return 'YELLOW' // Crafting
      return 'RED' // Waiting / No Rate
    }
    return 'NONE' // Active but Good Stock
  } else {
    // Inactive - User preference: "If replenish setting is not even on, it should show gray."
    // And "Red is requested... if not open replenish setting... show gray".
    // So Inactive is ALWAYS Gray.
    return 'GRAY'
  }
}

function shouldShowTriangle(task) {
  const type = getTaskStatusType(task)
  return type !== 'NONE'
}

function getStatusColor(task) {
  const type = getTaskStatusType(task)
  switch (type) {
    case 'YELLOW': return '#e6a23c'
    case 'RED': return '#f56c6c'
    case 'GRAY': return '#909399'
    default: return 'transparent'
  }
}

// Sort tasks based on status priority
const sortedAutoCraftTasks = computed(() => {
  const tasks = [...autoCraftTasks.value]
  // Priority: Yellow(0) -> Red(1) -> None(2) -> Gray(3)
  const priority = { 'YELLOW': 0, 'RED': 1, 'NONE': 2, 'GRAY': 3 }
  
  return tasks.sort((a, b) => {
    const typeA = getTaskStatusType(a)
    const typeB = getTaskStatusType(b)
    if (priority[typeA] !== priority[typeB]) {
      return priority[typeA] - priority[typeB]
    }
    // Secondary sort by itemId
    return (a.itemId || '').localeCompare(b.itemId || '')
  })
})
// Keeping getStockColor for detail dialog reuse or refactoring it there
function getStockColor(task) {
   const stock = inventoryIndex.value[task.itemId] || 0
   if (stock < task.minThreshold) return 'var(--secondary-color)'
   if (stock > task.maxThreshold) return 'var(--accent-color)'
   return 'var(--text-color)'
}

async function handleFetchCraftables() {
  try { await _fetchCraftables() } catch (err) { ElMessage.error(err.message) }
}
async function handleFinishWizard() {
  try { await _finishWizard() } catch (err) { ElMessage.error(err.message) }
}
async function handleSaveTaskThresholds() {
  try { await _saveTaskThresholds(); ElMessage.success('UPDATED') } catch (err) { ElMessage.error(err.message) }
}
async function handleDeleteTask(itemId) {
  try { await _deleteTask(itemId); ElMessage.success('DELETED'); setDetailVisible(false) } catch (err) { ElMessage.error(err.message) }
}
async function onActiveChange(task, value) {
  try { await _handleTaskActiveChange(task, value) } catch (err) { ElMessage.error(err.message) }
}

// Proxies
const updateWizardVisible = (v) => setWizardVisible(v)
const updateCraftableQuery = (v) => setCraftableQuery(v)
const updateMinThreshold = (v) => setMinThreshold(v)
const updateMaxThreshold = (v) => setMaxThreshold(v)
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
    
    // 2. Save thresholds
    await _saveTaskThresholds()
    
    // 3. Update active state if changed
    if (localForm.value.active !== detailTask.value.isActive) {
      await _handleTaskActiveChange(detailTask.value, localForm.value.active)
    }
    
    ElMessage.success(t('CRAFT.SAVE') + ' ' + 'OK') // detailed message if needed
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
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 15px;
}

.task-card {
  background: var(--surface-color);
  border: 2px solid var(--border-color);
  display: flex;
  flex-direction: column;
  cursor: pointer;
  transition: all 0.1s;
  box-shadow: 4px 4px 0 var(--border-color);
  aspect-ratio: 1; /* Make the whole card square */
  overflow: hidden; /* Clip the corner triangle */
  
  &:hover {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 var(--border-color);
  }
  
  &:active {
    transform: translate(4px, 4px);
    box-shadow: 0 0 0 var(--border-color);
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
  width: 65%;
  height: 65%;
}

.status-triangle {
  position: absolute;
  top: -18px;
  right: -18px;
  width: 36px;
  height: 36px;
  transform: rotate(45deg);
  border-bottom: 2px solid #000; /* Crisp diagonal border */
  box-shadow: -1px 1px 0 rgba(0,0,0,0.1);
  z-index: 1; /* Ensure it sits below text but above card bg */
}

.overlay-count {
  position: absolute;
  bottom: 8px;
  right: 3px;
  font-family: var(--font-nums);
  font-size: 1.4rem;
  color: #ffffff; /* White text */
  text-align: right;
  z-index: 2;
  pointer-events: none;
  line-height: 1;
  /* Brown stroke using text-shadow - Thickened */
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

/* FAB */
.brutalist-fab {
  position: fixed;
  right: 40px;
  bottom: 40px;
  width: 60px;
  height: 60px;
  background: var(--primary-color);
  color: #000;
  border: 3px solid var(--border-color);
  font-size: 2rem;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 6px 6px 0 rgba(0,0,0,0.2);
  transition: all 0.2s;
  
  &:hover {
    transform: translate(3px, 3px);
    box-shadow: 3px 3px 0 rgba(0,0,0,0.2);
  }

  &:active {
    transform: translate(6px, 6px);
    box-shadow: 0 0 0 rgba(0,0,0,0.2);
  }
}

/* Detail simplified */
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
  border-top: 2px dashed var(--secondary-color); /* Changed to dashed and possibly yellow if secondary is yellow, or explicit color */
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
  border: 2px solid var(--border-color) !important; /* 2px Black border to match ON button */
}

.b-summary { display: flex; gap: 15px; align-items: center; background: var(--bg-color); border: 2px solid var(--border-color); padding: 10px; margin-bottom: 10px; }

/* Wizard dialog styles */
.b-toolbar {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 15px;
  
  .el-input { flex: 1; }
}

.bg-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 4px;
  max-height: 40vh;
  overflow-y: auto;
  padding: 2px; /* Creates space for edge cell outlines */
  background: #cbccd4; /* Gap color = inner grid lines */
  border: 2px solid #f2f2f2; /* Outer edge border */
}

.bg-grid .grid-cell {
  aspect-ratio: 1;
  background: #adb0c4;
  border: none;
  box-shadow: inset 0 2px 0 #9a9fb4; /* Top shadow for 3D depth */
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.05s;
  padding: 2px;
  
  &:hover {
    background: #a6bedc;
    outline: 2px solid #daffff;
    z-index: 1;
  }
  
  &.selected {
    background: #a6bedc;
    outline: 2px solid #daffff;
    z-index: 1;
  }
}

.b-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  
  .actions {
    display: flex;
    gap: 10px;
  }
}

.b-inputs {
  display: flex;
  gap: 20px;
  
  .inp-grp {
    flex: 1;
    label { display: block; font-weight: bold; margin-bottom: 5px; }
  }
}

.centered-header {
  text-align: center;
  border-bottom: 2px dashed #e6a23c !important; /* Match bottom separator exactly: 2px dashed yellow */
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
    border: 2px solid var(--border-color) !important; /* Unified 2px black border */
    font-weight: bold;
  }
  
  /* ON button in dialog: shadow + press animation */
  .b-switch-rect {
    margin-left: 0;
    padding: 0 15px;
    min-width: 55px; /* Prevent size change between ON/OFF */
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
  
  /* Override Save button to match border thickness */
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

/* HIde close button for detail dialog */
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
