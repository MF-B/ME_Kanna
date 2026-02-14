<template>
  <section class="panel-section">
    <div class="panel-title autocraft-title">
      <span>自动合成任务</span>
      <span class="panel-subtitle">实时监控库存阈值并级联合成</span>
    </div>

    <el-row :gutter="20">
      <el-col
        v-for="task in autoCraftTasks"
        :key="task.itemId"
        :xs="24" :sm="12" :md="8" :lg="6"
        style="margin-bottom: 20px;"
      >
        <el-card class="autocraft-card" shadow="hover" @click="openTaskDetail(task)">
          <div class="autocraft-card-header">
            <ItemIcon :item-id="task.itemId" />
            <div class="autocraft-card-title">
              <div class="name">{{ displayItemName(task.itemId, task.itemName) }}</div>
              <div class="id">{{ task.itemId }}</div>
            </div>
            <el-tag :type="task.isActive ? 'success' : 'info'" effect="dark" size="small">
              {{ task.isActive ? 'ACTIVE' : 'PAUSED' }}
            </el-tag>
          </div>
          <div class="autocraft-card-body">
            <div class="autocraft-stat">
              <span>当前库存</span>
              <strong>{{ formatCompact(inventoryIndex[task.itemId] || 0) }}</strong>
            </div>
            <div class="autocraft-stat">
              <span>目标阈值</span>
              <strong>{{ task.minThreshold }} / {{ task.maxThreshold }}</strong>
            </div>
          </div>
          <div class="autocraft-card-footer">
            <el-switch
              :model-value="task.isActive"
              inline-prompt
              active-text="ON"
              inactive-text="OFF"
              @change="(value) => onActiveChange(task, value)"
              @click.stop
            />
            <div class="autocraft-actions">
              <span class="autocraft-footer-hint">点击编辑阈值</span>
              <el-button link type="danger" @click.stop="handleDeleteTask(task.itemId)">删除</el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty
      v-if="autoCraftTasks.length === 0"
      description="还没有自动合成任务，点击右下角 + 创建"
    />

    <teleport to="body">
      <el-button class="fab-button" type="primary" @click="openWizard">
        <span class="fab-icon">+</span>
      </el-button>
    </teleport>

    <el-dialog
      :model-value="wizardVisible"
      @update:model-value="updateWizardVisible"
      width="780px"
      class="autocraft-dialog"
      :append-to-body="true"
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-title">自动合成配置向导</div>
          <div class="dialog-subtitle">为你的工厂建立智能库存阈值</div>
        </div>
      </template>

      <el-steps :active="wizardStep - 1" finish-status="success" align-center>
        <el-step title="选择物品" />
        <el-step title="阈值设置" />
      </el-steps>

      <div class="wizard-body" v-if="wizardStep === 1">
        <div class="wizard-toolbar">
          <el-input :model-value="craftableQuery" @update:model-value="updateCraftableQuery" placeholder="搜索可合成物品" clearable />
          <el-button :loading="craftablesLoading" @click="handleFetchCraftables">刷新</el-button>
        </div>
        <div class="craftable-list">
          <el-scrollbar height="280px">
            <div
              v-for="item in filteredCraftables"
              :key="item.itemId"
              class="craftable-row"
              :class="{ selected: selectedCraftable && selectedCraftable.itemId === item.itemId }"
              @click="selectCraftable(item)"
            >
              <ItemIcon :item-id="item.itemId" />
              <div class="craftable-meta">
                <div class="name">{{ displayItemName(item.itemId, item.itemName) }}</div>
                <div class="id">{{ item.itemId }}</div>
              </div>
            </div>
          </el-scrollbar>
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
.autocraft-title {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.panel-subtitle {
  font-size: 0.85rem;
  text-transform: none;
  letter-spacing: 0.5px;
  color: #7b86a9;
}

.autocraft-card {
  background: linear-gradient(140deg, rgba(20, 24, 35, 0.95), rgba(14, 18, 28, 0.85));
  border: 1px solid rgba(79, 110, 255, 0.25);
  border-radius: 16px;
  color: #eef2ff;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.35);
  cursor: pointer;
}

.autocraft-card-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 12px;
  align-items: center;
}

.autocraft-card-title .name {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5ff;
}

.autocraft-card-title .id {
  font-size: 0.75rem;
  color: #8b93aa;
}

.autocraft-card-body {
  margin-top: 14px;
  display: grid;
  gap: 8px;
}

.autocraft-stat {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: #c3c9dd;
}

.autocraft-stat strong {
  color: #e6ecff;
}

.autocraft-card-footer {
  margin-top: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.autocraft-footer-hint {
  font-size: 0.75rem;
  color: #7f88a6;
}

.fab-button {
  position: fixed;
  right: 36px;
  bottom: 36px;
  width: 56px;
  height: 56px;
  border: none;
  background: radial-gradient(circle at top, #58f0c2, #1b7f6a);
  box-shadow: 0 12px 28px rgba(27, 127, 106, 0.4);
  color: #0d1216;
  z-index: 10;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.fab-icon {
  font-size: 2rem;
  line-height: 1;
  font-weight: 300;
  margin-top: -4px;
}

.fab-button:hover {
  transform: translateY(-4px) scale(1.05);
  box-shadow: 0 16px 36px rgba(27, 127, 106, 0.5);
}

.autocraft-dialog :deep(.el-dialog__body),
.autocraft-detail-dialog :deep(.el-dialog__body) {
  padding-top: 14px;
}

.dialog-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.dialog-title {
  font-size: 1.3rem;
  font-weight: 700;
  color: #e6f6ff;
}

.dialog-subtitle {
  font-size: 0.85rem;
  color: #7c8ab0;
}

.wizard-body {
  margin-top: 20px;
}

.wizard-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.craftable-list {
  border: 1px solid rgba(96, 115, 168, 0.25);
  border-radius: 12px;
  background: rgba(11, 15, 24, 0.7);
}

.craftable-row {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 10px 14px;
  cursor: pointer;
  border-bottom: 1px solid rgba(96, 115, 168, 0.15);
}

.craftable-row:last-child {
  border-bottom: none;
}

.craftable-row.selected {
  background: rgba(61, 214, 165, 0.15);
}

.craftable-meta .name {
  font-weight: 600;
  color: #eff3ff;
}

.craftable-meta .id {
  font-size: 0.75rem;
  color: #8b93aa;
}

.wizard-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: rgba(12, 16, 25, 0.7);
  border-radius: 12px;
  border: 1px solid rgba(96, 115, 168, 0.3);
}

.wizard-summary .name {
  font-weight: 600;
  color: #f1f5ff;
}

.wizard-summary .id {
  font-size: 0.8rem;
  color: #8b93aa;
}

.threshold-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-top: 18px;
}

.threshold-item {
  padding: 12px 14px;
  background: rgba(12, 16, 25, 0.7);
  border-radius: 12px;
  border: 1px solid rgba(96, 115, 168, 0.3);
}

.threshold-item .label {
  font-size: 0.8rem;
  color: #8b93aa;
  margin-bottom: 8px;
}

.threshold-hint {
  margin-top: 12px;
  font-size: 0.8rem;
  color: #7c8ab0;
}

.tree-panel {
  margin-top: 16px;
}

.detail-actions {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .fab-button {
    right: 18px;
    bottom: 18px;
  }
}
</style>
