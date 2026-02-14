<template>
  <section class="panel-section">
    <div class="ae2-header-bar" style="margin-bottom: 20px;">
      <div class="autocraft-title">
        <span class="name">自动合成任务管理</span>
        <span class="panel-subtitle">实时监控库存阈值并级联合成</span>
      </div>
    </div>

    <div class="task-list-container">
      <div
        v-for="task in autoCraftTasks"
        :key="task.itemId"
        class="ae2-slot-row task-item"
        @click="openTaskDetail(task)"
      >
        <ItemIcon :item-id="task.itemId" />
        <div class="task-meta">
          <div class="name">{{ displayItemName(task.itemId, task.itemName) }}</div>
          <div class="id">{{ task.itemId }}</div>
        </div>
        
        <div class="task-stats-group">
          <div class="stat-pill">
            <span class="label">库存</span>
            <span class="value">{{ formatCompact(inventoryIndex[task.itemId] || 0) }}</span>
          </div>
          <div class="stat-pill">
            <span class="label">阈值</span>
            <span class="value">{{ task.minThreshold }} / {{ task.maxThreshold }}</span>
          </div>
        </div>

        <div class="task-actions" @click.stop>
          <el-switch
            :model-value="task.isActive"
            inline-prompt
            active-text="ON"
            inactive-text="OFF"
            @change="(value) => onActiveChange(task, value)"
          />
          <el-button link type="danger" @click="handleDeleteTask(task.itemId)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <el-empty
      v-if="autoCraftTasks.length === 0"
      description="无活跃合成任务"
    />

    <teleport to="body">
      <el-button class="fab-button ae2-side-button" type="primary" @click="openWizard">
        <span class="fab-icon">+</span>
      </el-button>
    </teleport>

    <!-- 向导弹窗重构 -->
    <el-dialog
      :model-value="wizardVisible"
      @update:model-value="updateWizardVisible"
      width="640px"
      class="autocraft-dialog"
      :append-to-body="true"
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-title">新建合成任务</div>
        </div>
      </template>

      <div class="wizard-body" v-if="wizardStep === 1">
        <div class="wizard-toolbar" style="margin-bottom: 15px;">
          <el-input :model-value="craftableQuery" @update:model-value="updateCraftableQuery" placeholder="搜索物品..." clearable />
          <el-button :loading="craftablesLoading" @click="handleFetchCraftables">刷新</el-button>
        </div>
        
        <div class="ae2-slot-grid craftable-grid-viewport">
          <div
            v-for="item in filteredCraftables"
            :key="item.itemId"
            class="ae2-slot-cell"
            :class="{ 'is-selected': selectedCraftable && selectedCraftable.itemId === item.itemId }"
            @click="selectCraftable(item)"
            :title="displayItemName(item.itemId, item.itemName)"
          >
            <ItemIcon :item-id="item.itemId" />
          </div>
        </div>
        <div class="selection-info" v-if="selectedCraftable" style="margin-top: 10px;">
          当前选择: <span class="text-purple">{{ displayItemName(selectedCraftable.itemId, selectedCraftable.itemName) }}</span>
        </div>
      </div>

      <div class="wizard-body" v-else-if="wizardStep === 2">
        <div class="wizard-summary">
          <ItemIcon :item-id="selectedCraftable?.itemId" />
          <div>
            <div class="name">{{ displayItemName(selectedCraftable?.itemId, selectedCraftable?.itemName) }}</div>
            <div class="id">{{ selectedCraftable?.itemId }}</div>
          </div>
        </div>
        <div class="threshold-grid">
          <div class="threshold-item">
            <div class="label">最低触发阈值</div>
            <el-input-number :model-value="minThreshold" @update:model-value="updateMinThreshold" :min="1" :step="1" />
          </div>
          <div class="threshold-item">
            <div class="label">补货目标阈值</div>
            <el-input-number :model-value="maxThreshold" @update:model-value="updateMaxThreshold" :min="1" :step="1" />
          </div>
        </div>
        <div class="threshold-hint">默认建议: 最低 64，目标 256</div>
      </div>

      <template #footer>
        <div class="wizard-footer">
          <el-button @click="closeWizard">取消</el-button>
          <el-button v-if="wizardStep > 1" @click="prevWizardStep">上一步</el-button>
          <el-button
            v-if="wizardStep < 2"
            type="primary"
            :disabled="wizardStep === 1 && !selectedCraftable"
            @click="nextWizardStep"
          >
            下一步
          </el-button>
          <el-button
            v-if="wizardStep === 2"
            type="primary"
            :loading="recipeLoading"
            @click="handleFinishWizard"
          >
            创建任务
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      :model-value="detailVisible"
      @update:model-value="updateDetailVisible"
      width="860px"
      class="autocraft-detail-dialog"
      :append-to-body="true"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-title">任务详情</div>
          <div class="dialog-subtitle">可直接调整阈值并查看依赖树</div>
        </div>
      </template>

      <div v-if="detailTask" class="detail-editor">
        <div class="wizard-summary">
          <ItemIcon :item-id="detailTask.itemId" />
          <div>
            <div class="name">{{ displayItemName(detailTask.itemId, detailTask.itemName) }}</div>
            <div class="id">{{ detailTask.itemId }}</div>
          </div>
        </div>

        <div class="threshold-grid">
          <div class="threshold-item">
            <div class="label">最低触发阈值</div>
            <el-input-number :model-value="detailMinThreshold" @update:model-value="updateDetailMinThreshold" :min="1" :step="1" />
          </div>
          <div class="threshold-item">
            <div class="label">补货目标阈值</div>
            <el-input-number :model-value="detailMaxThreshold" @update:model-value="updateDetailMaxThreshold" :min="1" :step="1" />
          </div>
        </div>

        <div class="detail-actions">
          <el-button type="primary" :loading="detailSaving" @click="handleSaveTaskThresholds">保存阈值</el-button>
        </div>
      </div>

      <div v-if="detailTask && detailTask.recipeSnapshot" class="tree-panel">
        <AutoCraftTree
          :node="detailTask.recipeSnapshot"
          :inventory-index="inventoryIndex"
          :task-index="taskIndex"
        />
      </div>
      <el-empty v-else description="还没有合成配方数据" />
    </el-dialog>
  </section>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useSystemStore } from '../../stores/systemStore'
import ItemIcon from '../ItemIcon.vue'
import AutoCraftTree from '../AutoCraftTree.vue'

const systemStore = useSystemStore()

// 使用 storeToRefs 获取响应式状态
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

// 直接从 store 获取方法，引用极其稳定
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
  try {
    await _fetchCraftables()
  } catch (err) {
    ElMessage.error(err.message || '刷新列表失败')
  }
}

async function handleFinishWizard() {
  try {
    await _finishWizard()
  } catch (err) {
    ElMessage.error(err.message || '创建任务失败')
  }
}

async function handleSaveTaskThresholds() {
  try {
    await _saveTaskThresholds()
    ElMessage.success('阈值已更新')
  } catch (err) {
    ElMessage.error(err.message || '保存失败')
  }
}

async function handleDeleteTask(itemId) {
  try {
    await _deleteTask(itemId)
    ElMessage.success('任务已删除')
  } catch (err) {
    ElMessage.error(err.message || '删除失败')
  }
}

async function onActiveChange(task, value) {
  try {
    await _handleTaskActiveChange(task, value)
  } catch (err) {
    ElMessage.error(err.message || '更新状态失败')
  }
}

function updateCraftableQuery(value) {
  setCraftableQuery(value)
}

function updateMinThreshold(value) {
  setMinThreshold(value)
}

function updateMaxThreshold(value) {
  setMaxThreshold(value)
}

function updateDetailMinThreshold(value) {
  setDetailMinThreshold(value)
}

function updateDetailMaxThreshold(value) {
  setDetailMaxThreshold(value)
}

function updateWizardVisible(value) {
  setWizardVisible(value)
}

function updateDetailVisible(value) {
  setDetailVisible(value)
}
</script>

<style scoped>
.task-list-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.task-item {
  cursor: pointer;
  transition: border-color 0.2s;
}

.task-item:hover {
  border-color: #ffffff;
}

.task-meta {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.task-meta .name {
  font-weight: bold;
  color: var(--ae2-text);
}

.task-meta .id {
  font-size: 0.75rem;
  color: #666;
}

.task-stats-group {
  display: flex;
  gap: 15px;
  margin: 0 20px;
}

.stat-pill {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-pill .label {
  font-size: 0.7rem;
  color: #777;
  text-transform: uppercase;
}

.stat-pill .value {
  font-weight: bold;
  color: var(--ae2-purple);
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.craftable-grid-viewport {
  height: 300px;
  overflow-y: auto;
  align-content: flex-start;
}

.autocraft-title {
  display: flex;
  flex-direction: column;
}

.autocraft-title .name {
  font-weight: bold;
  font-size: 1.1rem;
}

.panel-subtitle {
  font-size: 0.8rem;
  color: #666;
}

.fab-button {
  position: fixed;
  right: 36px;
  bottom: 36px;
  z-index: 100;
}
</style>
