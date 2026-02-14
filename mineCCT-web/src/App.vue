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
            <SystemOverview
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
          </el-tab-pane>

          <el-tab-pane label="工厂面板" name="factory">
            <FactoryPanel :connected="connected" :factories="factories" @command="handleCommand" />
          </el-tab-pane>

          <el-tab-pane label="库存控制" name="inventory-control">
            <AutoCraftPanel />
          </el-tab-pane>
        </el-tabs>
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
  background-clip: text;
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

.tree-panel {
  margin-top: 16px;
}

.detail-actions {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
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
