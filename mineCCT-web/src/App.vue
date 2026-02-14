<template>
  <el-config-provider :locale="zhCn">
    <div class="ae2-sidebar-container">
      <!-- 左侧 AE2 侧边栏 -->
      <aside class="ae2-sidebar">
        <el-tooltip content="监控面板" placement="right" popper-class="ae2-tooltip">
          <div 
            class="ae2-side-button" 
            :class="{ 'is-active': activeTab === 'monitor' }"
            @click="activeTab = 'monitor'"
          >
            <el-icon><Monitor /></el-icon>
          </div>
        </el-tooltip>

        <el-tooltip content="工厂面板" placement="right" popper-class="ae2-tooltip">
          <div 
            class="ae2-side-button" 
            :class="{ 'is-active': activeTab === 'factory' }"
            @click="activeTab = 'factory'"
          >
            <el-icon><OfficeBuilding /></el-icon>
          </div>
        </el-tooltip>

        <el-tooltip content="库存控制" placement="right" popper-class="ae2-tooltip">
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
            <h1 class="header-title">
              {{ activeTab === 'monitor' ? 'ME 监控终端' : activeTab === 'factory' ? 'ME 工厂终端' : 'ME 合成终端' }}
            </h1>
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
/* 样式已迁移至 assets/base.css 和 assets/main.css，此处保持精简 */
body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}
</style>
