<template>
  <el-card
    class="factory-card"
    :class="{ 'is-active': factory.isActive }"
    shadow="hover"
  >
    <div class="card-header">
      <div class="title-group">
        <ItemIcon :item-id="factory.itemId" />
        <div class="title-text">
          <span class="factory-name">{{ factory.name }}</span>
          <span class="factory-id">ID: {{ factory.id }}</span>
        </div>
      </div>
      
      <el-tag
        :type="factory.isActive ? 'success' : 'info'"
        effect="dark"
        size="small"
        class="status-tag"
      >
        {{ factory.isActive ? 'RUNNING' : 'STOPPED' }}
      </el-tag>
    </div>

    <div class="card-body">
      <div class="data-row">
        <div class="label">
          <el-icon><Odometer /></el-icon> 速率
        </div>
        <div class="value-group">
          <div class="big-value" :class="getRateColor(factory.prodRate)">
            {{ formatNumber(factory.prodRate) }} <span class="unit">/h</span>
          </div>
        </div>
      </div>

      <el-divider class="custom-divider" />

      <div class="data-row">
        <div class="label">
          <el-icon><Box /></el-icon> 库存
        </div>
        <div class="value-group">
          <div class="mid-value">
            {{ formatCompact(factory.count) }}
          </div>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <span class="control-label">远程控制</span>
      <el-switch
        v-model="localActive"
        inline-prompt
        active-text="ON"
        inactive-text="OFF"
        style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949"
        @change="handleSwitchChange"
        :loading="loading"
      />
    </div>
  </el-card>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Odometer, Box } from '@element-plus/icons-vue'
import ItemIcon from './ItemIcon.vue' // 确保你目录下有这个组件

const props = defineProps(['factory'])
const emit = defineEmits(['command']) // 改名为 command 更语义化

const loading = ref(false)
const localActive = ref(props.factory.isActive)

// 监听后端状态变化，同步给开关
// 防止网页手动开关后，后端还没更新，导致开关状态不一致
watch(() => props.factory.isActive, (newVal) => {
  localActive.value = newVal
  loading.value = false // 收到后端更新，说明操作完成
})

// 处理开关点击
function handleSwitchChange(val) {
  loading.value = true // 开启加载状态，防止连点
  
  // val 为 true (开启) -> 发送 "start"
  // val 为 false (关闭) -> 发送 "stop"
  const action = val ? 'start' : 'stop'
  
  // 向上触发事件，由父组件通过 WebSocket 发送
  emit('command', { 
    target: props.factory.id, 
    action: action 
  })
  
  // 这里的 loading 不会立刻结束，而是等 watch 到 props 变化或者超时后结束
  setTimeout(() => { loading.value = false }, 2000)
}

// 数字格式化: 1,234
function formatNumber(num) {
  if (!num) return '0'
  return Math.floor(num).toLocaleString()
}

// 大数字缩略: 1.5k, 2.3M
function formatCompact(num) {
  if (!num) return '0'
  // 如果数字小于 10000，直接显示完整数字
  if (num < 10000) return num.toLocaleString()
  
  return Intl.NumberFormat('en-US', {
    notation: "compact",
    maximumFractionDigits: 1
  }).format(num)
}

// 简单的颜色逻辑
function getRateColor(rate) {
  if (!rate) return 'text-gray'
  if (rate > 0) return 'text-green'
  return 'text-gray'
}
</script>

<style scoped>
/* 字体引入 (推荐在 index.html 引入 Google Fonts) */
@import url('https://fonts.googleapis.com/css2?family=Oswald:wght@500;700&display=swap');

.factory-card {
  background: #1d1e1f;
  border: 1px solid #363637;
  color: #fff;
  transition: all 0.3s ease;
  margin-bottom: 20px;
  border-radius: 8px;
}

.factory-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.4);
  border-color: #606266;
}

/* 激活状态左边框 */
.factory-card.is-active {
  border-left: 5px solid #67c23a;
}

/* --- Header --- */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-text {
  display: flex;
  flex-direction: column;
}

.factory-name {
  font-weight: bold;
  font-size: 1.2em;
  color: #E5EAF3;
}

.factory-id {
  font-size: 0.8em;
  color: #909399;
  font-family: monospace;
}

/* --- Body --- */
.card-body {
  padding: 0 5px;
}

.data-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.label {
  color: #909399;
  font-size: 0.9em;
  display: flex;
  align-items: center;
  gap: 6px;
}

.value-group {
  text-align: right;
}

.big-value {
  font-family: 'Oswald', sans-serif;
  font-size: 2em;
  line-height: 1;
  font-weight: 500;
}

.mid-value {
  font-family: 'Oswald', sans-serif;
  font-size: 1.6em;
  color: #fff;
}

.unit {
  font-size: 0.5em;
  color: #606266;
  font-weight: normal;
}

.custom-divider {
  margin: 15px 0;
  border-top: 1px dashed #414243;
}

/* --- Colors --- */
.text-green { color: #67c23a; text-shadow: 0 0 15px rgba(103, 194, 58, 0.25); }
.text-gray { color: #606266; }

/* --- Footer --- */
.card-footer {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #303133;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.control-label {
  font-size: 0.85em;
  font-weight: bold;
  color: #909399;
}
</style>
