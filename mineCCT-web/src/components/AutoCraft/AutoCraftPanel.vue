<template>
  <div class="autocraft-panel">
    <div class="task-list-container">
      <div
        v-for="(task, index) in autoCraftTasks"
        :key="task.itemId"
        class="task-ticket"
        :class="{ pulse: !task.isActive }"
        @click="openTaskDetail(task)"
        :style="{ animationDelay: index * 0.1 + 's' }"
      >
        <div class="ticket-stub">
          <ItemIcon :item-id="task.itemId" class="task-icon" />
          <div class="ticket-id">#{{ index + 1 }}</div>
        </div>
        
        <div class="ticket-body">
          <div class="ticket-header">
            <h3>{{ displayItemName(task.itemId, task.itemName) }}</h3>
            <span class="item-id-code">{{ task.itemId }}</span>
          </div>
          
          <div class="ticket-stats">
            <div class="stat-box">
              <span class="lbl">STOCK</span>
              <span class="val">{{ formatCompact(inventoryIndex[task.itemId] || 0) }}</span>
            </div>
            <div class="stat-box">
              <span class="lbl">TARGET</span>
              <span class="val">{{ task.minThreshold }} / {{ task.maxThreshold }}</span>
            </div>
          </div>
        </div>

        <div class="ticket-actions" @click.stop>
          <div class="b-switch-mini" :class="{ active: task.isActive }" @click="onActiveChange(task, !task.isActive)"></div>
          <button class="b-btn-danger" @click="handleDeleteTask(task.itemId)">X</button>
        </div>
      </div>
    </div>

    <el-empty
      v-if="autoCraftTasks.length === 0"
      description="NO TASKS ACTIVE"
      class="brutalist-empty"
    />

    <!-- FAB -->
    <button class="brutalist-fab" @click="openWizard">
      +
    </button>

    <!-- Dialogs -->
    <el-dialog
      :model-value="wizardVisible"
      @update:model-value="updateWizardVisible"
      width="640px"
      class="brutalist-dialog"
      :append-to-body="true"
    >
      <template #header>
        <div class="b-dialog-header">NEW TASK</div>
      </template>

      <div v-if="wizardStep === 1">
        <div class="b-toolbar">
          <el-input :model-value="craftableQuery" @update:model-value="updateCraftableQuery" placeholder="SEARCH ITEM..." clearable />
          <el-button @click="handleFetchCraftables">REFRESH</el-button>
        </div>
        
        <div class="bg-grid">
          <div
            v-for="item in filteredCraftables"
            :key="item.itemId"
            class="grid-cell"
            :class="{ selected: selectedCraftable && selectedCraftable.itemId === item.itemId }"
            @click="selectCraftable(item)"
          >
            <ItemIcon :item-id="item.itemId" :count="inventoryIndex[item.itemId] || 0" />
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
             <label>MIN THRESHOLD</label>
             <el-input-number :model-value="minThreshold" @update:model-value="updateMinThreshold" :min="1" />
           </div>
           <div class="inp-grp">
             <label>MAX THRESHOLD</label>
             <el-input-number :model-value="maxThreshold" @update:model-value="updateMaxThreshold" :min="1" />
           </div>
        </div>
      </div>

      <template #footer>
        <div class="b-footer">
          <el-button @click="closeWizard">CANCEL</el-button>
          <div class="actions">
            <el-button v-if="wizardStep > 1" @click="prevWizardStep">BACK</el-button>
            <el-button
              v-if="wizardStep < 2"
              type="primary"
              :disabled="wizardStep === 1 && !selectedCraftable"
              @click="nextWizardStep"
            >NEXT</el-button>
            <el-button
              v-if="wizardStep === 2"
              type="primary"
              :loading="recipeLoading"
              @click="handleFinishWizard"
            >CREATE</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog
      :model-value="detailVisible"
      @update:model-value="updateDetailVisible"
      width="860px"
      class="brutalist-dialog"
      :append-to-body="true"
    >
       <!-- (Simplified for brevity, reusing logic) -->
       <template #header><div class="b-dialog-header">TASK DETAILS</div></template>
       
       <div v-if="detailTask" class="detail-content">
          <div class="b-summary">
             <ItemIcon :item-id="detailTask.itemId" />
             <div class="info">
                <h4>{{ displayItemName(detailTask.itemId, detailTask.itemName) }}</h4>
                <p>{{ detailTask.itemId }}</p>
             </div>
          </div>
          
          <div class="b-inputs">
             <div class="inp-grp">
               <label>MIN</label>
               <el-input-number :model-value="detailMinThreshold" @update:model-value="updateDetailMinThreshold" />
             </div>
             <div class="inp-grp">
               <label>MAX</label>
               <el-input-number :model-value="detailMaxThreshold" @update:model-value="updateDetailMaxThreshold" />
             </div>
             <el-button type="primary" :loading="detailSaving" @click="handleSaveTaskThresholds">SAVE</el-button>
          </div>

          <div v-if="detailTask && detailTask.recipeSnapshot" class="tree-view">
             <AutoCraftTree :node="detailTask.recipeSnapshot" :inventory-index="inventoryIndex" :task-index="taskIndex" />
          </div>
       </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useSystemStore } from '../../stores/systemStore'
import ItemIcon from '../ItemIcon.vue'
import AutoCraftTree from '../AutoCraftTree.vue'

const systemStore = useSystemStore()

const {
  autoCraftTasks,
  inventoryIndex,
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
  try { await _deleteTask(itemId); ElMessage.success('DELETED') } catch (err) { ElMessage.error(err.message) }
}
async function onActiveChange(task, value) {
  try { await _handleTaskActiveChange(task, value) } catch (err) { ElMessage.error(err.message) }
}

// Proxies
const updateCraftableQuery = (v) => setCraftableQuery(v)
const updateMinThreshold = (v) => setMinThreshold(v)
const updateMaxThreshold = (v) => setMaxThreshold(v)
const updateDetailMinThreshold = (v) => setDetailMinThreshold(v)
const updateDetailMaxThreshold = (v) => setDetailMaxThreshold(v)
const updateWizardVisible = (v) => setWizardVisible(v)
const updateDetailVisible = (v) => setDetailVisible(v)
</script>

<style scoped lang="scss">
.task-list-container {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.task-ticket {
  display: flex;
  background: var(--surface-color);
  border: 2px solid #fff;
  transition: all 0.2s;
  cursor: pointer;
  animation: slideIn 0.3s forwards;
  opacity: 0;
  transform: translateX(-20px);
  
  &:hover {
    transform: translateX(5px);
    box-shadow: 6px 6px 0 var(--primary-color);
  }
}

@keyframes slideIn {
  to { opacity: 1; transform: translateX(0); }
}

.ticket-stub {
  width: 60px;
  border-right: 2px dashed #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px;
  background: #000;
  
  .ticket-id {
    font-size: 0.8rem;
    color: #666;
    margin-top: 5px;
  }
}

.ticket-body {
  flex: 1;
  padding: 15px;
}

.ticket-header {
  margin-bottom: 10px;
  h3 { margin: 0; font-size: 1.1rem; }
  .item-id-code { font-size: 0.7rem; color: #888; font-family: monospace; }
}

.ticket-stats {
  display: flex;
  gap: 20px;
  
  .stat-box {
    display: flex;
    flex-direction: column;
    
    .lbl { font-size: 0.7rem; color: #666; }
    .val { font-weight: bold; }
  }
}

.ticket-actions {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 10px;
  gap: 10px;
  border-left: 2px solid #fff;
  background: #000;
}

.b-switch-mini {
  width: 30px;
  height: 15px;
  border: 2px solid #fff;
  background: #333;
  cursor: pointer;
  
  &.active { background: var(--accent-color); }
}

.b-btn-danger {
  background: transparent;
  border: 2px solid var(--secondary-color);
  color: var(--secondary-color);
  font-weight: bold;
  cursor: pointer;
  
  &:hover {
    background: var(--secondary-color);
    color: #000;
  }
}

.brutalist-fab {
  position: fixed;
  right: 40px;
  bottom: 40px;
  width: 60px;
  height: 60px;
  background: var(--primary-color);
  color: #000;
  border: 3px solid #fff;
  font-size: 2rem;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 6px 6px 0 #000;
  transition: all 0.2s;
  
  &:hover {
    transform: translate(-2px, -2px);
    box-shadow: 8px 8px 0 #000;
  }
}

/* Dialog partials */
.b-dialog-header {
  font-size: 1.5rem;
  font-weight: 900;
  border-bottom: 3px solid var(--primary-color);
  padding-bottom: 10px;
}

.b-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.bg-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 2px;
  background: #333;
  padding: 2px;
  border: 2px solid #fff;
  max-height: 400px;
  overflow-y: auto;
  
  .grid-cell {
    aspect-ratio: 1;
    background: #000;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    border: 1px solid transparent;
    
    &:hover { border-color: #fff; }
    &.selected { border-color: var(--primary-color); background: #333; }
  }
}

.b-summary {
  display: flex;
  gap: 15px;
  align-items: center;
  background: #000;
  border: 2px solid #fff;
  padding: 15px;
  margin-bottom: 20px;
  
  .info h4 { margin: 0; }
  .info p { margin: 0; color: #888; font-size: 0.8rem; }
}

.b-inputs {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
  
  .inp-grp { flex: 1; }
}

.b-footer {
  display: flex;
  justify-content: space-between;
}
</style>
