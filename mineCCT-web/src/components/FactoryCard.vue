<template>
  <el-card
    class="factory-card"
    :class="{ 'is-active': factory.isActive }"
    shadow="hover"
    @click="openSettings"
  >
    <div class="card-header">
      <div class="title-group">
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
      <div class="item-list" v-if="visibleItems.length">
        <div class="item-row" v-for="item in visibleItems" :key="item.itemId">
          <ItemIcon :item-id="item.itemId" />
          <div class="item-meta">
            <div class="item-id">{{ item.itemId }}</div>
            <div class="item-stats">
              {{ formatNumber(item.prodRate) }}/h · {{ formatCompact(item.count) }}
            </div>
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
        @click.stop
        :loading="loading"
      />
    </div>
  </el-card>

  <el-dialog
    v-model="dialogVisible"
    title="工厂显示配置"
    width="520px"
    :append-to-body="true"
  >
    <div class="name-panel">
      <div class="settings-title">工厂名称</div>
      <el-input
        v-model="localName"
        placeholder="输入自定义名称"
        @input="nameEditing = true"
        @change="commitName"
        @blur="commitName"
        @click.stop
      />
    </div>

    <div v-if="localItems.length" class="settings-panel">
      <div class="settings-title">显示配置 + 拖动排序</div>
      <div class="settings-group">
        <div
          class="settings-row"
          v-for="(item, index) in localItems"
          :key="item.itemId"
          draggable="true"
          @dragstart="handleDragStart(index)"
          @dragover.prevent="handleDragOver(index)"
          @drop="handleDrop"
        >
          <el-checkbox v-model="item.visible" @change="commitSettings"></el-checkbox>
          <span class="drag-handle">||</span>
          <span class="settings-id">{{ item.itemId }}</span>
        </div>
      </div>
    </div>
    <div v-else class="settings-empty">暂无产物数据</div>
  </el-dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import ItemIcon from './ItemIcon.vue' // 确保你目录下有这个组件

const props = defineProps(['factory'])
const emit = defineEmits(['command']) // 改名为 command 更语义化

const loading = ref(false)
const localActive = ref(props.factory.isActive)
const localItems = ref([])
const dialogVisible = ref(false)
const dragIndex = ref(null)
const localName = ref(props.factory?.name || '')
const nameEditing = ref(false)

// 监听后端状态变化，同步给开关
// 防止网页手动开关后，后端还没更新，导致开关状态不一致
watch(() => props.factory.isActive, (newVal) => {
  localActive.value = newVal
  loading.value = false // 收到后端更新，说明操作完成
})

watch(() => props.factory, (factory) => {
  const items = normalizeItems(factory?.items)
  localItems.value = sortItems(items).map((item) => ({ ...item }))
  if (!nameEditing.value && !dialogVisible.value) {
    localName.value = factory?.name || ''
  }
}, { deep: true, immediate: true })

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

function openSettings() {
  dialogVisible.value = true
  localName.value = props.factory?.name || ''
  nameEditing.value = false
}

function handleDragStart(index) {
  dragIndex.value = index
}

function handleDragOver(index) {
  if (dragIndex.value === null || dragIndex.value === index) return
  const updated = [...localItems.value]
  const [moved] = updated.splice(dragIndex.value, 1)
  updated.splice(index, 0, moved)
  localItems.value = updated
  dragIndex.value = index
}

function handleDrop() {
  dragIndex.value = null
  commitSettings()
}

function commitName() {
  const name = localName.value?.trim()
  if (!name) return
  emit('command', {
    target: props.factory.id,
    action: 'update_factory_name',
    name
  })
  nameEditing.value = false
}

const itemsList = computed(() => normalizeItems(props.factory?.items))
const sortedItems = computed(() => sortItems(itemsList.value))

const visibleItems = computed(() => {
  const list = sortedItems.value.filter((item) => item.visible)
  return list
})

function normalizeItems(items) {
  if (!items) return []
  const list = Array.isArray(items) ? items : Object.values(items)
  return list.map((item) => ({
    ...item,
    visible: item.visible !== false,
    order: Number.isFinite(item.order) ? item.order : 0
  }))
}

function sortItems(items) {
  const hasOrder = items.some((item) => item.order && item.order > 0)
  const byName = (a, b) => (a.itemId || '').localeCompare(b.itemId || '')
  if (!hasOrder) {
    return [...items].sort(byName)
  }
  return [...items].sort((a, b) => {
    const orderA = a.order && a.order > 0 ? a.order : Number.POSITIVE_INFINITY
    const orderB = b.order && b.order > 0 ? b.order : Number.POSITIVE_INFINITY
    if (orderA !== orderB) return orderA - orderB
    return byName(a, b)
  })
}

function commitSettings() {
  emit('command', {
    target: props.factory.id,
    action: 'update_factory_items',
    items: localItems.value.map((item, index) => ({
      itemId: item.itemId,
      visible: item.visible,
      order: index + 1
    }))
  })
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
  padding: 0 2px 4px;
}

.item-list {
  margin-top: 6px;
  display: grid;
  gap: 6px;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 6px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.04);
}

.item-row :deep(.item-icon-wrapper) {
  width: 28px;
  height: 28px;
  padding: 3px;
}

.item-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-id {
  font-size: 0.8em;
  color: #cdd0d6;
}

.item-stats {
  font-size: 0.75em;
  color: #909399;
}

.settings-panel {
  display: grid;
  gap: 12px;
}

.name-panel {
  display: grid;
  gap: 8px;
  margin-bottom: 16px;
}

.settings-title {
  font-size: 0.9em;
  color: #909399;
}

.settings-group {
  display: grid;
  gap: 10px;
}

.settings-row {
  display: grid;
  grid-template-columns: auto 28px 1fr;
  gap: 8px;
  align-items: center;
  cursor: grab;
}

.settings-row:active {
  cursor: grabbing;
}

.drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  border: 1px dashed rgba(255, 255, 255, 0.2);
  color: #909399;
  font-family: monospace;
  font-size: 0.75em;
  user-select: none;
}

.settings-id {
  font-size: 0.82em;
  color: #c0c4cc;
  word-break: break-all;
}

.settings-empty {
  color: #909399;
  font-size: 0.9em;
}

/* --- Colors --- */
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
