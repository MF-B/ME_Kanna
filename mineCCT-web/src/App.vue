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
      </main>

    </div>
  </el-config-provider>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import FactoryCard from './components/FactoryCard.vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { ElMessage } from 'element-plus'

// --- 状态 ---
const connected = ref(false)
const factories = ref([])

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
        
      }
    } catch (e) {
      console.error('Data parse error', e)
    }
  }
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
body {
  margin: 0;
  background-color: #0d0d0d;
  color: #e5eaf3;
  font-family: 'Inter', 'Helvetica Neue', sans-serif;
  min-height: 100vh;
}

.dashboard-container {
  padding: 20px 40px;
  max-width: 1600px;
  margin: 0 auto;
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
</style>
