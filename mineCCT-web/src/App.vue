<template>
  <el-config-provider :locale="zhCn">
    <div class="brutalist-container">
      <!-- Brutalist Sidebar -->
      <aside class="brutalist-sidebar" :class="{ collapsed: isCollapsed }">
        <div class="brand">
          <h1 v-if="!isCollapsed">ME<br><span class="highlight">Kanna</span></h1>
          <h1 v-else>M<span class="highlight">K</span></h1>
        </div>

        <nav class="brutalist-nav">
          <div 
            class="nav-item" 
            :class="{ active: activeTab === 'monitor' }"
            @click="activeTab = 'monitor'"
          >
            <span>01</span> <template v-if="!isCollapsed">{{ t('NAV.MONITOR') }}</template>
          </div>
          
          <div 
            class="nav-item" 
            :class="{ active: activeTab === 'factory' }"
            @click="activeTab = 'factory'"
          >
            <span>02</span> <template v-if="!isCollapsed">{{ t('NAV.FACTORY') }}</template>
          </div>

          <div 
            class="nav-item" 
            :class="{ active: activeTab === 'inventory-control' }"
            @click="activeTab = 'inventory-control'"
          >
            <span>03</span> <template v-if="!isCollapsed">{{ t('NAV.CRAFT') }}</template>
          </div>
        </nav>

        <div class="sidebar-footer">
          <div class="collapse-btn" @click="toggleSidebar">
            {{ isCollapsed ? '>>' : '<<' }}
          </div>
          <div v-if="!isCollapsed" class="footer-content">
            <div class="lang-toggle" @click="toggleLocale">
              <span :class="{ active: locale === 'en' }">EN</span> / 
              <span :class="{ active: locale === 'zh' }">CN</span>
            </div>
            <p>SYS.VER.2.1</p>
          </div>
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
import { useI18n } from './composables/useI18n'

const activeTab = ref('monitor')
const systemStore = useSystemStore()
const { t, locale, toggleLocale } = useI18n()
const isCollapsed = ref(false)

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
}

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
  loadAutoCraftTasks,
  fetchCraftables
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
    await Promise.all([
      loadAutoCraftTasks(),
      fetchCraftables()
    ])
  } catch (err) {
    console.error(err)
  }

  // Responsive sidebar: Auto-collapse on small screens
  const mediaQuery = window.matchMedia('(max-width: 1000px)')
  const handleResize = (e) => {
    isCollapsed.value = e.matches
  }
  // Initial check
  if (mediaQuery.matches) isCollapsed.value = true
  // Listen
  mediaQuery.addEventListener('change', handleResize)
})
</script>

<style scoped lang="scss">
.brutalist-container {
  display: flex;
  width: 100vw;
  height: 100vh;
}

.brand {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px dashed var(--border-color);
  
  h1 {
    font-size: 3rem;
    line-height: 1;
    margin-bottom: 0;
    color: var(--text-color);
    
    .highlight {
      color: var(--surface-color);
      background: var(--secondary-color);
      padding: 0 5px;
    }
  }
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  background: var(--bg-color);
  color: var(--text-color);
  font-weight: bold;
  border: 2px solid var(--border-color);
  
  &.online {
    background: var(--accent-color);
    color: #fff;
    box-shadow: 4px 4px 0 rgba(0,0,0,0.1);
  }
}

.brutalist-nav {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.nav-item {
  font-size: 1.2rem;
  font-weight: bold;
  cursor: pointer;
  padding: 15px;
  background: var(--surface-color);
  border: 2px solid var(--border-color);
  box-shadow: 4px 4px 0 var(--border-color);
  transition: all 0.1s;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-color);
  
  span {
    font-size: 1rem;
    color: #666;
  }
  
  &:hover {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 var(--border-color);
    background: var(--primary-color);
    
    span { color: #000; }
  }
  
  &.active {
    background: var(--primary-color);
    color: #000;
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 var(--border-color);
    border: 2px solid var(--border-color);
    
    span { color: #000; }
  }
  
  &:active {
    transform: translate(4px, 4px);
    box-shadow: none;
  }
}

.sidebar-footer {
  margin-top: auto;
  font-size: 0.8rem;
  color: #666;
  border-top: 2px solid var(--border-color);
  padding-top: 10px;
  display: flex;
  flex-direction: column;
  align-items: center; /* Center everything */
  gap: 10px;
  
  .lang-toggle {
    cursor: pointer;
    margin-bottom: 5px;
    font-weight: bold;
    
    span {
      padding: 0 5px;
      &.active {
        background: var(--text-color);
        color: var(--surface-color);
      }
    }
    
    &:hover {
      color: var(--primary-color);
    }
  }
}

.content-viewport {
  padding: 2rem;
  height: 100%;
  overflow-y: auto;
}

/* Collapsible styles */
.brutalist-sidebar {
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  width: 300px; /* Default width */
  overflow: hidden;
  
  &.collapsed {
    width: 80px;
    padding: 20px 10px;
    
    .brand h1 { font-size: 2rem; text-align: center; }
    .nav-item { justify-content: center; padding: 15px 5px; }
    .nav-item span { font-size: 1.2rem; color: var(--text-color); }
    .sidebar-footer { flex-direction: column; align-items: center; }
  }
}

.footer-content {
  margin-top: 10px;
  text-align: center;
}

.collapse-btn {
  cursor: pointer;
  font-weight: bold;
  padding: 5px 10px;
  border: 2px solid var(--border-color);
  background: var(--surface-color);
  display: inline-block;
  margin-bottom: 5px;
  
  &:hover {
    background: var(--primary-color);
  }
}
</style>
