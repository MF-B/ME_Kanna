<template>
  <el-config-provider :locale="zhCn">
    <div class="brutalist-container">
      <!-- Brutalist Sidebar -->
      <aside class="brutalist-sidebar">
        <div class="brand">
          <h1>MINE<br><span class="highlight">CCT</span></h1>
          <div class="status-badge" :class="{ online: connected }">
            {{ connected ? 'ONLINE' : 'OFFLINE' }}
          </div>
        </div>

        <nav class="brutalist-nav">
          <div 
            class="nav-item" 
            :class="{ active: activeTab === 'monitor' }"
            @click="activeTab = 'monitor'"
          >
            <span>01</span> MONITOR
          </div>
          
          <div 
            class="nav-item" 
            :class="{ active: activeTab === 'factory' }"
            @click="activeTab = 'factory'"
          >
            <span>02</span> FACTORY
          </div>

          <div 
            class="nav-item" 
            :class="{ active: activeTab === 'inventory-control' }"
            @click="activeTab = 'inventory-control'"
          >
            <span>03</span> CRAFT
          </div>
        </nav>

        <div class="sidebar-footer">
          <p>SYS.VER.2.0</p>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="brutalist-main">
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
  onOpen: () => ElMessage.success('SYSTEM ONLINE'),
  onClose: () => {
    systemStore.resetFactories()
    ElMessage.warning('SYSTEM OFFLINE')
  }
})

const handleCommand = (payload) => {
  if (!send(payload)) {
    ElMessage.error('CONNECTION_ERROR')
    return
  }
}

onMounted(async () => {
  connect()
  try {
    await loadAutoCraftTasks()
  } catch (err) {
    console.error(err)
  }
})
</script>

<style scoped lang="scss">
.brutalist-container {
  display: flex;
  width: 100vw;
  height: 100vh;
}

.brand {
  margin-bottom: 3rem;
  
  h1 {
    font-size: 3rem;
    line-height: 1;
    margin-bottom: 1rem;
    
    .highlight {
      color: var(--primary-color);
      background: #000;
      padding: 0 5px;
    }
  }
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #333;
  color: #fff;
  font-weight: bold;
  border: 2px solid #fff;
  
  &.online {
    background: var(--accent-color);
    color: #000;
    box-shadow: 4px 4px 0 rgba(0,0,0,0.5);
  }
}

.brutalist-nav {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.nav-item {
  font-size: 1.5rem;
  font-weight: bold;
  cursor: pointer;
  padding: 10px;
  border: 2px solid transparent;
  transition: all 0.2s;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 10px;
  
  span {
    font-size: 1rem;
    color: #666;
  }
  
  &:hover {
    padding-left: 20px;
    color: var(--primary-color);
    
    span { color: var(--primary-color); }
  }
  
  &.active {
    background: var(--primary-color);
    color: #000;
    border: 3px solid #fff;
    box-shadow: 6px 6px 0 #000;
    
    span { color: #000; }
  }
}

.sidebar-footer {
  margin-top: auto;
  font-size: 0.8rem;
  color: #666;
  border-top: 2px solid #333;
  padding-top: 10px;
}

.content-viewport {
  padding: 2rem;
  height: 100%;
  overflow-y: auto;
}
</style>
