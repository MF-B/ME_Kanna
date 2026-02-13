<template>
  <el-config-provider :locale="zhCn">
    <div class="dashboard-container">

      <header class="dashboard-header">
        <div class="header-left-group">
          <div class="brand">
            <h1>MF_B 的机器</h1>
          </div>
        </div>

        <div class="header-controls">
          <div class="status-bar">
            <el-badge is-dot :type="connected ? 'success' : 'danger'">
              <span class="status-text" :class="{ online: connected }">
                {{ connected ? 'System Online' : 'Offline' }}
              </span>
            </el-badge>
          </div>
        </div>
      </header>

      <main>
        <el-tabs v-model="activeTab" class="dashboard-tabs">
          <el-tab-pane label="监控面板" name="monitor">
            <section class="panel-section">
              <div class="panel-title">AE2 监控面板</div>

              <el-row :gutter="20">
                <el-col :xs="24" :md="12" :lg="10">
                  <el-card class="energy-card" shadow="hover">
                    <div class="energy-header">
                      <div class="energy-title">能量储备</div>
                      <el-tag size="small" effect="dark" :type="energyPercent < 20 ? 'danger' : 'success'">
                        {{ Math.floor(energyPercent) }}%
                      </el-tag>
                    </div>

                    <el-progress
                      :percentage="energyPercent"
                      :stroke-width="10"
                      :color="energyColor"
                      :show-text="false"
                    />

                    <div class="energy-meta">
                      <div class="energy-value">
                        {{ formatCompact(systemStatus.energyStored) }} / {{ formatCompact(systemStatus.energyMax) }} AE
                      </div>
                      <div class="energy-updated" v-if="systemStatus.lastUpdated">
                        更新: {{ formatTime(systemStatus.lastUpdated) }}
                      </div>
                    </div>

                    <div class="energy-grid">
                      <div class="energy-stat">
                        <div class="label">输入均值</div>
                        <div class="value">{{ formatRate(systemStatus.averageEnergyInput) }}</div>
                      </div>
                      <div class="energy-stat">
                        <div class="label">消耗速率</div>
                        <div class="value">{{ formatRate(systemStatus.energyUsage) }}</div>
                      </div>
                      <div class="energy-stat">
                        <div class="label">净变化</div>
                        <div class="value" :class="netRateClass">{{ formatRate(systemStatus.netEnergyRate, true) }}</div>
                      </div>
                    </div>
                  </el-card>
                </el-col>

                <el-col :xs="24" :md="12" :lg="14">
                  <el-card class="storage-card" shadow="hover">
                    <div class="energy-header">
                      <div class="energy-title">库存总览</div>
                      <el-tag size="small" effect="dark" :type="storagePercent < 80 ? 'success' : 'warning'">
                        {{ Math.floor(storagePercent) }}%
                      </el-tag>
                    </div>

                    <div class="storage-split" role="img" aria-label="内部与外部存储容量占比">
                      <div class="storage-segment internal" :style="{ width: `${storageInternalRatio}%` }">
                        <div class="storage-fill" :style="{ width: `${storageInternalUsage}%` }"></div>
                      </div>
                      <div class="storage-segment external" :style="{ width: `${storageExternalRatio}%` }">
                        <div class="storage-fill" :style="{ width: `${storageExternalUsage}%` }"></div>
                      </div>
                    </div>

                    <div class="storage-legend">
                      <div class="legend-item">
                        <span class="legend-swatch internal"></span>
                        内部 {{ formatCompact(systemStatus.storage.itemTotal) }}
                      </div>
                      <div class="legend-item">
                        <span class="legend-swatch external"></span>
                        外部 {{ formatCompact(systemStatus.storage.itemExternalTotal) }}
                      </div>
                    </div>

                    <div class="energy-meta">
                      <div class="energy-value">
                        {{ formatCompact(storageTotalUsed) }} / {{ formatCompact(storageTotalCapacity) }} 物品存储
                      </div>
                      <div class="energy-updated" v-if="systemStatus.lastUpdated">
                        更新: {{ formatTime(systemStatus.lastUpdated) }}
                      </div>
                    </div>

                    <div class="storage-grid">
                      <div class="storage-block">
                        <div class="block-title">物品存储</div>
                        <div class="block-row">已用 {{ formatCompact(systemStatus.storage.itemUsed) }} / 总计 {{ formatCompact(systemStatus.storage.itemTotal) }}</div>
                        <div class="block-row muted">外部 {{ formatCompact(systemStatus.storage.itemExternalUsed) }} / {{ formatCompact(systemStatus.storage.itemExternalTotal) }}</div>
                      </div>
                      <div class="storage-block">
                        <div class="block-title">流体存储</div>
                        <div class="block-row">已用 {{ formatCompact(systemStatus.storage.fluidUsed) }} / 总计 {{ formatCompact(systemStatus.storage.fluidTotal) }}</div>
                        <div class="block-row muted">外部 {{ formatCompact(systemStatus.storage.fluidExternalUsed) }} / {{ formatCompact(systemStatus.storage.fluidExternalTotal) }}</div>
                      </div>
                    </div>
                  </el-card>
                </el-col>
              </el-row>
            </section>
          </el-tab-pane>

          <el-tab-pane label="工厂面板" name="factory">
            <section class="panel-section">
              <div class="panel-title">工厂产能与库存</div>

              <el-row :gutter="20">
                <el-col
                  v-for="factory in factories"
                  :key="factory.id"
                  :xs="24" :sm="12" :md="8" :lg="6"
                  style="margin-bottom: 20px;"
                >
                  <FactoryCard
                    :factory="factory"
                    @command="handleCommand"
                  />
                </el-col>
              </el-row>

              <el-empty
                v-if="factories.length === 0 && connected"
                description="等待 AE 网络数据上报..."
              />

              <el-empty
                v-if="!connected"
                description="与 Go 后端断开连接，正在重试..."
                :image-size="100"
              />
            </section>
          </el-tab-pane>

          <el-tab-pane label="库存控制" name="inventory-control">
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
                        v-model="task.isActive"
                        inline-prompt
                        active-text="ON"
                        inactive-text="OFF"
                        @click.stop
                      />
                      <span class="autocraft-footer-hint">点击查看依赖树</span>
                    </div>
                  </el-card>
                </el-col>
              </el-row>

              <el-empty
                v-if="autoCraftTasks.length === 0"
                description="还没有自动合成任务，点击右下角 + 创建"
              />
            </section>
          </el-tab-pane>
        </el-tabs>
      </main>

      <el-button
        v-if="activeTab === 'inventory-control'"
        class="fab-button"
        type="primary"
        @click="openWizard"
      >
        <span class="fab-icon">+</span>
      </el-button>

      <el-dialog
        v-model="wizardVisible"
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
          <el-step title="级联原料" />
        </el-steps>

        <div class="wizard-body" v-if="wizardStep === 1">
          <div class="wizard-toolbar">
            <el-input v-model="craftableQuery" placeholder="搜索可合成物品" clearable />
            <el-button :loading="craftablesLoading" @click="fetchCraftables">刷新</el-button>
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
              <el-input-number v-model="minThreshold" :min="1" :step="1" />
            </div>
            <div class="threshold-item">
              <div class="label">补货目标阈值</div>
              <el-input-number v-model="maxThreshold" :min="1" :step="1" />
            </div>
          </div>
          <div class="threshold-hint">默认建议: 最低 64，目标 256</div>
        </div>

        <div class="wizard-body" v-else>
          <div class="wizard-summary">
            <div>
              <div class="name">级联设置</div>
              <div class="id">是否也为原料自动建立阈值任务?</div>
            </div>
          </div>
          <div class="cascade-list">
            <el-checkbox-group v-model="cascadeSelection">
              <div
                v-for="material in cascadeOptions"
                :key="material.itemId"
                class="cascade-row"
              >
                <el-checkbox :label="material.itemId">
                  <div class="cascade-item">
                    <ItemIcon :item-id="material.itemId" />
                    <div>
                      <div class="name">{{ displayItemName(material.itemId, material.itemName) }}</div>
                      <div class="id">默认 {{ defaultMinThreshold }} / {{ defaultMaxThreshold }}</div>
                    </div>
                  </div>
                </el-checkbox>
              </div>
            </el-checkbox-group>
          </div>
        </div>

        <template #footer>
          <div class="wizard-footer">
            <el-button @click="closeWizard">取消</el-button>
            <el-button v-if="wizardStep > 1" @click="prevWizardStep">上一步</el-button>
            <el-button
              v-if="wizardStep < 3"
              type="primary"
              :disabled="wizardStep === 1 && !selectedCraftable"
              :loading="recipeLoading"
              @click="nextWizardStep"
            >
              下一步
            </el-button>
            <el-button
              v-else
              type="primary"
              @click="finishWizard"
            >
              创建任务
            </el-button>
          </div>
        </template>
      </el-dialog>

      <el-dialog
        v-model="detailVisible"
        width="860px"
        class="autocraft-detail-dialog"
        :append-to-body="true"
      >
        <template #header>
          <div class="dialog-header">
            <div class="dialog-title">依赖树</div>
            <div class="dialog-subtitle">实时库存 vs 阈值目标</div>
          </div>
        </template>

        <div v-if="detailTask && detailTask.recipeSnapshot" class="tree-panel">
          <AutoCraftTree
            :node="detailTask.recipeSnapshot"
            :inventory-index="inventoryIndex"
            :task-index="taskIndex"
          />
        </div>
        <el-empty v-else description="还没有合成配方数据" />
      </el-dialog>

    </div>
  </el-config-provider>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import FactoryCard from './components/FactoryCard.vue'
import ItemIcon from './components/ItemIcon.vue'
import AutoCraftTree from './components/AutoCraftTree.vue'
import { useItemNames } from './composables/useItemNames'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { ElMessage } from 'element-plus'

// --- 状态 ---
const connected = ref(false)
const factories = ref([])
const activeTab = ref('monitor')
const autoCraftTasks = ref([])

const wizardVisible = ref(false)
const wizardStep = ref(1)
const craftables = ref([])
const craftablesLoading = ref(false)
const craftableQuery = ref('')
const selectedCraftable = ref(null)
const minThreshold = ref(64)
const maxThreshold = ref(256)
const recipeTree = ref(null)
const recipeLoading = ref(false)
const cascadeOptions = ref([])
const cascadeSelection = ref([])

const detailVisible = ref(false)
const detailTask = ref(null)
const { names: itemNameMap, ensureName } = useItemNames()

const systemStatus = ref({
  energyStored: 0,
  energyMax: 1,
  energyUsage: 0,
  averageEnergyInput: 0,
  netEnergyRate: 0,
  lastUpdated: 0,
  storage: {
    itemTotal: 0,
    itemUsed: 0,
    itemExternalTotal: 0,
    itemExternalUsed: 0,
    fluidTotal: 0,
    fluidUsed: 0,
    fluidExternalTotal: 0,
    fluidExternalUsed: 0
  }
})

let socket = null

const defaultMinThreshold = 64
const defaultMaxThreshold = 256

// --- WebSocket 逻辑 ---
const connectWS = () => {
  let host = window.location.hostname
  if (host.includes(':') && !host.startsWith('[')) {
    host = `[${host}]`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const wsUrl = `${protocol}://${host}:8080/ws/web`

  console.log(`Connecting to Backend: ${wsUrl}`)

  if (socket) socket.close()

  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    console.log('WS Connected')
    connected.value = true
    ElMessage.success('已连接到控制中心')
  }

  socket.onclose = () => {
    console.log('WS Disconnected')
    connected.value = false
    factories.value = []
    setTimeout(connectWS, 3000)
  }

  socket.onerror = (err) => {
    console.error('WS Error:', err)
    socket.close()
  }

  socket.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data)
      if (payload.type === 'update') {
        // 更新工厂列表
        factories.value = payload.data
        if (payload.system) {
          systemStatus.value = payload.system
        }
      }
    } catch (e) {
      console.error('Data parse error', e)
    }
  }
}

const inventoryIndex = computed(() => {
  const index = {}
  factories.value.forEach((factoryData) => {
    const factoryItems = factoryData?.items || {}
    Object.values(factoryItems).forEach((factoryItem) => {
      if (factoryItem?.itemId) {
        index[factoryItem.itemId] = factoryItem.count || 0
      }
    })
  })
  return index
})

const taskIndex = computed(() => {
  const index = {}
  autoCraftTasks.value.forEach((task) => {
    index[task.itemId] = task
  })
  return index
})

const energyPercent = computed(() => {
  if (!systemStatus.value.energyMax) return 0
  const p = (systemStatus.value.energyStored / systemStatus.value.energyMax) * 100
  return Math.min(Math.max(p, 0), 100)
})

const energyColor = computed(() => energyPercent.value < 20 ? '#f56c6c' : '#3dd6a5')

const storagePercent = computed(() => {
  if (!storageTotalCapacity.value) return 0
  const p = (storageTotalUsed.value / storageTotalCapacity.value) * 100
  return Math.min(Math.max(p, 0), 100)
})

const storageColor = computed(() => storagePercent.value > 90 ? '#ff6b6b' : '#5d8aff')

const storageTotalCapacity = computed(() => {
  return (systemStatus.value.storage.itemTotal || 0) + (systemStatus.value.storage.itemExternalTotal || 0)
})

const storageTotalUsed = computed(() => {
  return (systemStatus.value.storage.itemUsed || 0) + (systemStatus.value.storage.itemExternalUsed || 0)
})

const storageInternalRatio = computed(() => {
  if (!storageTotalCapacity.value) return 50
  return (systemStatus.value.storage.itemTotal / storageTotalCapacity.value) * 100
})

const storageExternalRatio = computed(() => {
  if (!storageTotalCapacity.value) return 50
  return (systemStatus.value.storage.itemExternalTotal / storageTotalCapacity.value) * 100
})

const storageInternalUsage = computed(() => {
  if (!systemStatus.value.storage.itemTotal) return 0
  return Math.min((systemStatus.value.storage.itemUsed / systemStatus.value.storage.itemTotal) * 100, 100)
})

const storageExternalUsage = computed(() => {
  if (!systemStatus.value.storage.itemExternalTotal) return 0
  return Math.min((systemStatus.value.storage.itemExternalUsed / systemStatus.value.storage.itemExternalTotal) * 100, 100)
})

const netRateClass = computed(() => {
  if (systemStatus.value.netEnergyRate > 0) return 'text-green'
  if (systemStatus.value.netEnergyRate < 0) return 'text-red'
  return 'text-gray'
})

function formatCompact(num) {
  if (num === null || num === undefined) return '0'
  return Intl.NumberFormat('en-US', {
    notation: "compact",
    maximumFractionDigits: 1
  }).format(num)
}

function formatRate(num, signed) {
  if (num === null || num === undefined) return '0 AE/t'
  const value = signed && num > 0 ? `+${formatCompact(num)}` : formatCompact(num)
  return `${value} AE/t`
}

function formatTime(epochSeconds) {
  if (!epochSeconds) return '--:--:--'
  const date = new Date(epochSeconds * 1000)
  return date.toLocaleTimeString('zh-CN', { hour12: false })
}

function getApiBase() {
  const host = window.location.hostname
  const protocol = window.location.protocol
  return `${protocol}//${host}:8080`
}

const filteredCraftables = computed(() => {
  const query = craftableQuery.value.trim().toLowerCase()
  if (!query) return craftables.value
  return craftables.value.filter((craftableItem) => {
    const itemIdKeyword = (craftableItem.itemId || '').toLowerCase()
    const itemNameKeyword = displayItemName(craftableItem.itemId, craftableItem.itemName).toLowerCase()
    return itemIdKeyword.includes(query) || itemNameKeyword.includes(query)
  })
})

async function fetchCraftables() {
  craftablesLoading.value = true
  try {
    const res = await fetch(`${getApiBase()}/autocraft/craftables`)
    const data = await res.json()
    const craftableList = Array.isArray(data) ? data : (data.items || [])
    craftables.value = craftableList.map((craftableEntry) => {
      if (typeof craftableEntry === 'string') {
        return { itemId: craftableEntry, itemName: craftableEntry }
      }
      return {
        itemId: craftableEntry.itemId,
        itemName: craftableEntry.itemName || craftableEntry.itemId
      }
    }).filter((craftableEntry) => craftableEntry.itemId)

    craftables.value.forEach((craftableEntry) => {
      ensureName(craftableEntry.itemId)
    })
  } catch (err) {
    craftables.value = []
    ElMessage.error('获取可合成列表失败')
  } finally {
    craftablesLoading.value = false
  }
}

function openWizard() {
  wizardVisible.value = true
  wizardStep.value = 1
  craftableQuery.value = ''
  selectedCraftable.value = null
  minThreshold.value = defaultMinThreshold
  maxThreshold.value = defaultMaxThreshold
  recipeTree.value = null
  cascadeOptions.value = []
  cascadeSelection.value = []
  fetchCraftables()
}

function closeWizard() {
  wizardVisible.value = false
}

function selectCraftable(item) {
  selectedCraftable.value = item
  ensureName(item.itemId)
}

function prevWizardStep() {
  if (wizardStep.value > 1) {
    wizardStep.value -= 1
  }
}

async function nextWizardStep() {
  if (wizardStep.value === 1) {
    wizardStep.value = 2
    return
  }
  if (wizardStep.value === 2) {
    recipeLoading.value = true
    try {
      const itemId = selectedCraftable.value?.itemId
      const res = await fetch(`${getApiBase()}/autocraft/recipe?itemId=${encodeURIComponent(itemId)}`)
      const data = await res.json()
      recipeTree.value = data
      ensureName(itemId)
      cascadeOptions.value = collectRecipeLeaves(data, itemId)
      cascadeSelection.value = cascadeOptions.value.map((materialEntry) => materialEntry.itemId)
      wizardStep.value = 3
    } catch (err) {
      ElMessage.error('获取配方失败')
    } finally {
      recipeLoading.value = false
    }
  }
}

function finishWizard() {
  const main = selectedCraftable.value
  if (!main) return

  upsertTask({
    itemId: main.itemId,
    itemName: displayItemName(main.itemId, main.itemName),
    minThreshold: minThreshold.value,
    maxThreshold: maxThreshold.value,
    isActive: true,
    recipeSnapshot: recipeTree.value
  })

  cascadeSelection.value.forEach((itemId) => {
    const node = findRecipeNode(recipeTree.value, itemId)
    upsertTask({
      itemId,
      itemName: displayItemName(itemId, node?.itemName || itemId),
      minThreshold: defaultMinThreshold,
      maxThreshold: defaultMaxThreshold,
      isActive: true,
      recipeSnapshot: node || null
    })
  })

  wizardVisible.value = false
}

function openTaskDetail(task) {
  detailTask.value = task
  detailVisible.value = true
}

function upsertTask(task) {
  const index = autoCraftTasks.value.findIndex((t) => t.itemId === task.itemId)
  if (index >= 0) {
    autoCraftTasks.value[index] = { ...autoCraftTasks.value[index], ...task }
  } else {
    autoCraftTasks.value = [...autoCraftTasks.value, task]
  }

  ensureName(task.itemId)
}

function collectRecipeLeaves(node, rootId) {
  const list = []
  const seen = new Set()
  const walk = (recipeNode) => {
    if (!recipeNode) return
    const children = recipeNode.children || []
    if (!children.length) {
      if (recipeNode.itemId && recipeNode.itemId !== rootId && !seen.has(recipeNode.itemId)) {
        seen.add(recipeNode.itemId)
        ensureName(recipeNode.itemId)
        list.push({ itemId: recipeNode.itemId, itemName: displayItemName(recipeNode.itemId, recipeNode.itemName || recipeNode.itemId) })
      }
      return
    }
    children.forEach(walk)
  }
  walk(node)
  return list
}

function findRecipeNode(node, itemId) {
  if (!node) return null
  if (node.itemId === itemId) return node
  const children = node.children || []
  for (const childNode of children) {
    const found = findRecipeNode(childNode, itemId)
    if (found) return found
  }
  return null
}

function displayItemName(itemId, fallbackName) {
  if (!itemId) return fallbackName || ''
  ensureName(itemId)
  return itemNameMap.value[itemId] || fallbackName || itemId
}

// --- 发送指令 ---
const handleCommand = (payload) => {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    ElMessage.error('网络未连接，无法发送指令')
    return
  }
  socket.send(JSON.stringify(payload))
  console.log('Sent Command:', payload)
}

onMounted(() => {
  document.documentElement.classList.add('dark')
  connectWS()
})

onUnmounted(() => {
  if (socket) socket.close()
})
</script>

<style>
/* 全局样式 */
@import url('https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;600;700&display=swap');

body {
  margin: 0;
  background: radial-gradient(circle at top, rgba(61, 214, 165, 0.18), transparent 55%),
    linear-gradient(180deg, #0b0c0f 0%, #0d0f14 100%);
  color: #e5eaf3;
  font-family: 'Space Grotesk', 'Helvetica Neue', sans-serif;
  min-height: 100vh;
}

.dashboard-container {
  padding: 20px 40px;
  max-width: 1600px;
  margin: 0 auto;
}

.dashboard-tabs {
  margin-top: 10px;
}

.panel-section {
  padding: 8px 0 10px;
  animation: fade-up 0.6s ease;
}

.panel-title {
  font-size: 1.1rem;
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: 16px;
  color: #c9d6ff;
}

/* Header 布局优化 */
.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  border-bottom: 1px solid #333;
  padding-bottom: 20px;
  flex-wrap: wrap; /* 允许小屏幕换行 */
  gap: 20px;
}

.header-left-group {
  display: flex;
  align-items: center;
  gap: 40px;
  flex-wrap: wrap;
}

h1 {
  margin: 0;
  font-size: 1.8rem;
  background: linear-gradient(90deg, #409eff, #67c23a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-transform: uppercase;
  letter-spacing: 2px;
  font-weight: 800;
}

.energy-card,
.storage-card {
  background: rgba(18, 20, 26, 0.92);
  border: 1px solid rgba(61, 214, 165, 0.25);
  color: #e5eaf3;
  border-radius: 14px;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.35);
}

.storage-card {
  border-color: rgba(93, 138, 255, 0.3);
}

.energy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.energy-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #f0f5ff;
}

.energy-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  color: #aab2c8;
  font-size: 0.9rem;
}

.energy-value {
  font-size: 1rem;
  font-weight: 600;
  color: #eef2ff;
}

.energy-updated {
  font-size: 0.8rem;
  color: #7f89a5;
}

.energy-grid {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
}

.energy-stat {
  background: rgba(15, 17, 22, 0.6);
  border: 1px solid rgba(120, 128, 162, 0.18);
  border-radius: 10px;
  padding: 12px;
}

.energy-stat .label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: #8b93aa;
}

.energy-stat .value {
  font-size: 1.05rem;
  font-weight: 600;
  margin-top: 6px;
}

.text-green {
  color: #3dd6a5;
}

.text-red {
  color: #ff6b6b;
}

.text-gray {
  color: #7c869e;
}

.storage-grid {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.storage-split {
  display: flex;
  height: 12px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(93, 138, 255, 0.2);
  margin-bottom: 10px;
}

.storage-segment {
  position: relative;
  display: flex;
  align-items: stretch;
}

.storage-segment.internal {
  background: rgba(61, 214, 165, 0.25);
}

.storage-segment.external {
  background: rgba(93, 138, 255, 0.25);
}

.storage-fill {
  height: 100%;
  background: linear-gradient(90deg, rgba(61, 214, 165, 0.95), rgba(61, 214, 165, 0.65));
}

.storage-segment.external .storage-fill {
  background: linear-gradient(90deg, rgba(93, 138, 255, 0.95), rgba(93, 138, 255, 0.65));
}

.storage-legend {
  display: flex;
  gap: 16px;
  font-size: 0.8rem;
  color: #aab2c8;
  margin-bottom: 6px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend-swatch {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  display: inline-block;
}

.legend-swatch.internal {
  background: #3dd6a5;
}

.legend-swatch.external {
  background: #5d8aff;
}

.storage-block {
  background: rgba(15, 17, 22, 0.6);
  border: 1px solid rgba(93, 138, 255, 0.2);
  border-radius: 10px;
  padding: 12px;
}

.storage-block .block-title {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: #9aa3c4;
  margin-bottom: 8px;
}

.storage-block .block-row {
  font-size: 0.9rem;
  color: #eef2ff;
  margin-bottom: 6px;
}

.storage-block .block-row.muted {
  color: #7e88a7;
}

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

.autocraft-dialog .el-dialog__body,
.autocraft-detail-dialog .el-dialog__body {
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

.cascade-list {
  margin-top: 16px;
  max-height: 260px;
  overflow: auto;
  border: 1px solid rgba(96, 115, 168, 0.25);
  border-radius: 12px;
  padding: 8px 12px;
}

.cascade-row {
  padding: 8px 0;
  border-bottom: 1px solid rgba(96, 115, 168, 0.15);
}

.cascade-row:last-child {
  border-bottom: none;
}

.cascade-item {
  display: flex;
  gap: 12px;
  align-items: center;
}

.tree-panel {
  margin-top: 16px;
}

@media (max-width: 768px) {
  .dashboard-container {
    padding: 16px 18px;
  }

  .fab-button {
    right: 18px;
    bottom: 18px;
  }
}

/* 状态灯样式 */
.status-text {
  font-weight: bold;
  margin-left: 8px;
  font-size: 0.9rem;
  color: #f56c6c;
  transition: color 0.3s;
}
.status-text.online {
  color: #67c23a;
}

@keyframes fade-up {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
