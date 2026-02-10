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
                        {{ energyPercent.toFixed(1) }}%
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
                        {{ storagePercent.toFixed(1) }}%
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
                        <div class="block-row muted">可用 内部 {{ formatCompact(systemStatus.storage.itemAvailable) }} / 外部 {{ formatCompact(systemStatus.storage.itemExternalAvailable) }}</div>
                      </div>
                      <div class="storage-block">
                        <div class="block-title">流体存储</div>
                        <div class="block-row">已用 {{ formatCompact(systemStatus.storage.fluidUsed) }} / 总计 {{ formatCompact(systemStatus.storage.fluidTotal) }}</div>
                        <div class="block-row muted">外部 {{ formatCompact(systemStatus.storage.fluidExternalUsed) }} / {{ formatCompact(systemStatus.storage.fluidExternalTotal) }}</div>
                        <div class="block-row muted">可用 内部 {{ formatCompact(systemStatus.storage.fluidAvailable) }} / 外部 {{ formatCompact(systemStatus.storage.fluidExternalAvailable) }}</div>
                      </div>
                      <div class="storage-block">
                        <div class="block-title">化学存储</div>
                        <div class="block-row">已用 {{ formatCompact(systemStatus.storage.chemicalUsed) }} / 总计 {{ formatCompact(systemStatus.storage.chemicalTotal) }}</div>
                        <div class="block-row muted">外部 {{ formatCompact(systemStatus.storage.chemicalExternalUsed) }} / {{ formatCompact(systemStatus.storage.chemicalExternalTotal) }}</div>
                        <div class="block-row muted">可用 内部 {{ formatCompact(systemStatus.storage.chemicalAvailable) }} / 外部 {{ formatCompact(systemStatus.storage.chemicalExternalAvailable) }}</div>
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
        </el-tabs>
      </main>

    </div>
  </el-config-provider>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import FactoryCard from './components/FactoryCard.vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { ElMessage } from 'element-plus'

// --- 状态 ---
const connected = ref(false)
const factories = ref([])
const activeTab = ref('monitor')

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
    itemAvailable: 0,
    itemExternalTotal: 0,
    itemExternalUsed: 0,
    itemExternalAvailable: 0,
    fluidTotal: 0,
    fluidUsed: 0,
    fluidAvailable: 0,
    fluidExternalTotal: 0,
    fluidExternalUsed: 0,
    fluidExternalAvailable: 0,
    chemicalTotal: 0,
    chemicalUsed: 0,
    chemicalAvailable: 0,
    chemicalExternalTotal: 0,
    chemicalExternalUsed: 0,
    chemicalExternalAvailable: 0
  }
})

let socket = null

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
