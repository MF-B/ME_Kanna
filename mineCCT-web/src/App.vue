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
              <div class="panel-title">AE2 能源监控</div>

              <el-row :gutter="20">
                <el-col :xs="24" :md="16" :lg="12">
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
  lastUpdated: 0
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

.energy-card {
  background: rgba(18, 20, 26, 0.92);
  border: 1px solid rgba(61, 214, 165, 0.25);
  color: #e5eaf3;
  border-radius: 14px;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.35);
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
