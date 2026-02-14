<template>
  <el-config-provider :locale="zhCn">
    <div class="ae2-sidebar-container">
      <!-- 左侧 AE2 侧边栏 -->
      <aside class="ae2-sidebar">
        <el-tooltip content="监控面板" placement="right">
          <div 
            class="ae2-side-button" 
            :class="{ 'is-active': activeTab === 'monitor' }"
            @click="activeTab = 'monitor'"
          >
            <el-icon><Monitor /></el-icon>
          </div>
        </el-tooltip>

        <el-tooltip content="工厂面板" placement="right">
          <div 
            class="ae2-side-button" 
            :class="{ 'is-active': activeTab === 'factory' }"
            @click="activeTab = 'factory'"
          >
            <el-icon><OfficeBuilding /></el-icon>
          </div>
        </el-tooltip>

        <el-tooltip content="库存控制" placement="right">
          <div 
            class="ae2-side-button" 
            :class="{ 'is-active': activeTab === 'inventory-control' }"
            @click="activeTab = 'inventory-control'"
          >
            <el-icon><Box /></el-icon>
          </div>
        </el-tooltip>
      </aside>

      <!-- 右侧主面板区 -->
      <main class="ae2-main-content">
        <div class="mc-panel dashboard-main">
          <header class="dashboard-header">
            <div class="header-title">
              {{ activeTab === 'monitor' ? 'ME 监控终端' : activeTab === 'factory' ? 'ME 工厂终端' : 'ME 合成终端' }}
            </div>
            <div class="status-indicator">
              <span class="status-text" :class="{ online: connected }">
                {{ connected ? 'ONLINE' : 'OFFLINE' }}
              </span>
            </div>
          </header>

          <div class="content-viewport">
            <SystemOverview
              v-if="activeTab === 'monitor'"
              :system-status="systemStatus"
              :energy-percent="energyPercent"
              :energy-color="energyColor"
              :storage-percent="storagePercent"
              :storage-total-used="storageTotalUsed"
              :storage-total-capacity="storageTotalCapacity"
              :storage-internal-ratio="storageInternalRatio"
              :storage-external-ratio="storageExternalRatio"
              :storage-internal-usage="storageInternalUsage"
              :storage-external-usage="storageExternalUsage"
              :net-rate-class="netRateClass"
              :format-compact="formatCompact"
              :format-rate="formatRate"
              :format-time="formatTime"
            />

            <FactoryPanel 
              v-else-if="activeTab === 'factory'"
              :connected="connected" 
              :factories="factories" 
              @command="handleCommand" 
            />

            <AutoCraftPanel 
              v-else-if="activeTab === 'inventory-control'"
            />
          </div>
        </div>
      </main>
    </div>
  </el-config-provider>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { ElMessage } from 'element-plus'
import SystemOverview from './components/Dashboard/SystemOverview.vue'
import FactoryPanel from './components/Inventory/FactoryPanel.vue'
import AutoCraftPanel from './components/AutoCraft/AutoCraftPanel.vue'
import { storeToRefs } from 'pinia'
import { useSystemStore } from './stores/systemStore'
import { useWebSocket } from './composables/useWebSocket'

const activeTab = ref('monitor')

const systemStore = useSystemStore()

const {
  factories,
  systemStatus,
  energyPercent,
  energyColor,
  storagePercent,
  storageTotalCapacity,
  storageTotalUsed,
  storageInternalRatio,
  storageExternalRatio,
  storageInternalUsage,
  storageExternalUsage,
  netRateClass
} = storeToRefs(systemStore)

const {
  formatCompact,
  formatRate,
  formatTime,
  applyUpdatePayload,
  loadAutoCraftTasks
} = systemStore

const { connected, connect, send } = useWebSocket({
  onUpdate: applyUpdatePayload,
  onOpen: () => ElMessage.success('已连接到控制中心'),
  onClose: () => {
    systemStore.resetFactories()
  }
})

// --- 发送指令 ---
const handleCommand = (payload) => {
  if (!send(payload)) {
    ElMessage.error('网络未连接，无法发送指令')
    return
  }
}

onMounted(async () => {
  document.documentElement.classList.add('dark')
  connect()
  try {
    await loadAutoCraftTasks()
  } catch (err) {
    ElMessage.error(err.message || '加载自动合成任务失败')
  }
})
</script>

<style>
/* 彻底精简 App.vue 样式，仅保留必要的全局变量和基础重置 */
body {
  margin: 0;
  padding: 0;
  background-color: #1a1a1a;
  overflow: hidden; /* 防止出现双滚动条 */
}

/* 状态字体的微调 */
.status-text {
  font-weight: bold;
  font-size: 0.9rem;
  color: #f56c6c;
}
.status-text.online {
  color: #2e7d32;
}

/* 适配侧边栏布局 */
.ae2-sidebar-container {
  display: flex;
  width: 100vw;
  height: 100vh;
}

.ae2-main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 10px;
  overflow: hidden;
}
</style>

