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

    <el-button class="fab-button" type="primary" @click="openWizard">
      <span class="fab-icon">+</span>
    </el-button>

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
