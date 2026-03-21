import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useItemInfo } from './useItemInfo'

export function useSystemData() {
  const factories = ref([])
  const autoCraftTasks = ref([])

  const wizardVisible = ref(false)
  const wizardStep = ref(1)
  const craftables = ref([])
  const craftablesLoading = ref(false)
  const craftableQuery = ref('')
  const selectedCraftable = ref(null)
  const minThreshold = ref(64)
  const maxThreshold = ref(256)
  const recipeTree = ref(null)
  const recipeLoading = ref(false)

  const detailVisible = ref(false)
  const detailTask = ref(null)
  const detailMinThreshold = ref(64)
  const detailMaxThreshold = ref(256)
  const detailSaving = ref(false)

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
      itemExternalTotal: 0,
      itemExternalUsed: 0,
      fluidTotal: 0,
      fluidUsed: 0,
      fluidExternalTotal: 0,
      fluidExternalUsed: 0
    },
    inventory: {}
  })

  const { names: itemNameMap, ensureName } = useItemInfo()

  const defaultMinThreshold = 64
  const defaultMaxThreshold = 256

  function getApiBase() {
    const host = window.location.hostname
    const protocol = window.location.protocol
    return `${protocol}//${host}:8080`
  }

  function applyUpdatePayload(payload) {
    if (payload.type === 'update') {
      factories.value = payload.data || []
      if (payload.system) {
        systemStatus.value = payload.system
      }
    } else if (payload.type === 'craftables') {
      const list = payload.craftables || []
      craftables.value = list.map((craftableEntry) => ({
        itemId: craftableEntry.itemId,
        itemName: craftableEntry.itemName || craftableEntry.itemId,
        count: craftableEntry.count || 0
      })).filter((c) => c.itemId)

      craftables.value.forEach((c) => ensureName(c.itemId))
      if (craftablesLoading.value) craftablesLoading.value = false
    }
  }

  function formatCompact(num) {
    if (num === null || num === undefined) return '0'
    return Intl.NumberFormat('en-US', {
      notation: 'compact',
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

  const inventoryIndex = computed(() => {
    const systemInventory = systemStatus.value.inventory || {}
    if (Object.keys(systemInventory).length > 0) {
      return systemInventory
    }

    const index = {}
    factories.value.forEach((factoryData) => {
      const factoryItems = factoryData?.items || {}
      Object.values(factoryItems).forEach((factoryItem) => {
        if (factoryItem?.itemId) {
          index[factoryItem.itemId] = factoryItem.count || 0
        }
      })
    })
    return index
  })

  const taskIndex = computed(() => {
    const index = {}
    autoCraftTasks.value.forEach((task) => {
      index[task.itemId] = task
    })
    return index
  })

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
    const intTotal = systemStatus.value.storage.itemTotal || 0
    const extTotal = systemStatus.value.storage.itemExternalTotal || 0
    const intUsed = systemStatus.value.storage.itemUsed || 0
    const extUsed = systemStatus.value.storage.itemExternalUsed || 0
    // AE2 Storage Bus 可能报告 used > total，取 max 作为有效容量
    return Math.max(intTotal, intUsed) + Math.max(extTotal, extUsed)
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
    const total = systemStatus.value.storage.itemTotal
    if (!total) return 0
    return Math.min((systemStatus.value.storage.itemUsed / total) * 100, 100)
  })

  const storageExternalUsage = computed(() => {
    const total = systemStatus.value.storage.itemExternalTotal
    const used = systemStatus.value.storage.itemExternalUsed
    if (!total && !used) return 0
    // AE2 Storage Bus: used 可能远大于 total，此时显示 100%
    const effectiveTotal = Math.max(total, used) || 1
    return Math.min((used / effectiveTotal) * 100, 100)
  })

  const netRateClass = computed(() => {
    if (systemStatus.value.netEnergyRate > 0) return 'text-green'
    if (systemStatus.value.netEnergyRate < 0) return 'text-red'
    return 'text-gray'
  })

  const filteredCraftables = computed(() => {
    const query = craftableQuery.value.trim().toLowerCase()
    if (!query) return craftables.value
    return craftables.value.filter((craftableItem) => {
      const itemIdKeyword = (craftableItem.itemId || '').toLowerCase()
      const itemNameKeyword = displayItemName(craftableItem.itemId, craftableItem.itemName).toLowerCase()
      return itemIdKeyword.includes(query) || itemNameKeyword.includes(query)
    })
  })

  function displayItemName(itemId, fallbackName) {
    if (!itemId) return fallbackName || ''
    ensureName(itemId)
    return itemNameMap.value[itemId] || fallbackName || itemId
  }

  async function fetchCraftables() {
    craftablesLoading.value = true
    const abortController = new AbortController()
    const timeoutId = setTimeout(() => abortController.abort(), 8000)
    try {
      const res = await fetch(`${getApiBase()}/autocraft/craftables`, {
        signal: abortController.signal
      })
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }
      const data = await res.json()
      const craftableList = Array.isArray(data) ? data : (data.items || [])
      craftables.value = craftableList.map((craftableEntry) => {
        if (typeof craftableEntry === 'string') {
          return { itemId: craftableEntry, itemName: craftableEntry }
        }
        return {
          itemId: craftableEntry.itemId,
          itemName: craftableEntry.itemName || craftableEntry.itemId
        }
      }).filter((craftableEntry) => craftableEntry.itemId)

      craftables.value.forEach((craftableEntry) => {
        ensureName(craftableEntry.itemId)
      })
    } catch (err) {
      craftables.value = []
      if (err?.name === 'AbortError') {
        ElMessage.error('获取可合成列表超时')
      } else {
        ElMessage.error('获取可合成列表失败')
      }
    } finally {
      clearTimeout(timeoutId)
      craftablesLoading.value = false
    }
  }

  async function loadAutoCraftTasks() {
    try {
      const response = await fetch(`${getApiBase()}/autocraft/tasks`)
      const data = await response.json()
      const taskList = Array.isArray(data) ? data : (data.items || [])
      autoCraftTasks.value = taskList.map((entry) => ({
        itemId: entry.itemId,
        itemName: entry.itemName || entry.itemId,
        minThreshold: entry.minThreshold,
        maxThreshold: entry.maxThreshold,
        isActive: !!entry.isActive,
        recipeSnapshot: entry.recipeSnapshot || null
      }))

      autoCraftTasks.value.forEach((taskEntry) => ensureName(taskEntry.itemId))
    } catch (_error) {
      autoCraftTasks.value = []
    }
  }

  function openWizard() {
    wizardVisible.value = true
    wizardStep.value = 1
    craftableQuery.value = ''
    selectedCraftable.value = null
    minThreshold.value = defaultMinThreshold
    maxThreshold.value = defaultMaxThreshold
    recipeTree.value = null
    fetchCraftables()
  }

  function closeWizard() {
    wizardVisible.value = false
  }

  function selectCraftable(item) {
    selectedCraftable.value = item
    ensureName(item.itemId)
  }

  function prevWizardStep() {
    if (wizardStep.value > 1) {
      wizardStep.value -= 1
    }
  }

  function nextWizardStep() {
    if (wizardStep.value === 1) {
      wizardStep.value = 2
    }
  }

  function upsertTask(task) {
    const index = autoCraftTasks.value.findIndex((t) => t.itemId === task.itemId)
    if (index >= 0) {
      Object.assign(autoCraftTasks.value[index], task)
    } else {
      autoCraftTasks.value.push(task)
    }
    ensureName(task.itemId)
  }

  async function finishWizard() {
    const main = selectedCraftable.value
    if (!main) return

    recipeLoading.value = true
    try {
      const recipeResponse = await fetch(`${getApiBase()}/autocraft/recipe?itemId=${encodeURIComponent(main.itemId)}`)
      const recipeData = await recipeResponse.json()
      recipeTree.value = recipeData

      const taskPayload = {
        itemId: main.itemId,
        itemName: displayItemName(main.itemId, main.itemName),
        minThreshold: minThreshold.value,
        maxThreshold: maxThreshold.value,
        isActive: false,
        recipeSnapshot: recipeData
      }

      const response = await fetch(`${getApiBase()}/autocraft/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(taskPayload)
      })

      if (!response.ok) {
        ElMessage.error('创建任务失败')
        return
      }

      const savedTask = await response.json()
      upsertTask({
        itemId: savedTask.itemId,
        itemName: savedTask.itemName || savedTask.itemId,
        minThreshold: savedTask.minThreshold,
        maxThreshold: savedTask.maxThreshold,
        isActive: !!savedTask.isActive,
        recipeSnapshot: savedTask.recipeSnapshot || recipeData
      })

      wizardVisible.value = false
    } catch (_error) {
      ElMessage.error('创建任务失败')
    } finally {
      recipeLoading.value = false
    }
  }

  function openTaskDetail(task) {
    detailTask.value = { ...task }
    detailMinThreshold.value = task.minThreshold
    detailMaxThreshold.value = task.maxThreshold
    detailVisible.value = true
  }

  async function saveTaskThresholds() {
    if (!detailTask.value) return
    if (detailMinThreshold.value <= 0 || detailMaxThreshold.value < detailMinThreshold.value) {
      ElMessage.error('阈值不合法：目标阈值必须大于等于最低阈值')
      return
    }

    detailSaving.value = true
    const abortController = new AbortController()
    const timeoutId = setTimeout(() => abortController.abort(), 8000)
    try {
      const payload = {
        itemId: detailTask.value.itemId,
        itemName: detailTask.value.itemName,
        minThreshold: detailMinThreshold.value,
        maxThreshold: detailMaxThreshold.value,
        isActive: !!detailTask.value.isActive,
        recipeSnapshot: detailTask.value.recipeSnapshot || null
      }

      const response = await fetch(`${getApiBase()}/autocraft/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: abortController.signal
      })

      if (!response.ok) {
        let message = '保存阈值失败'
        try {
          const errorData = await response.json()
          if (errorData?.error) {
            message = `保存阈值失败: ${errorData.error}`
          }
        } catch (_ignore) { }
        ElMessage.error(message)
        return
      }

      const savedTask = await response.json()
      if (detailTask.value) {
        Object.assign(detailTask.value, {
          itemId: savedTask.itemId,
          itemName: savedTask.itemName || savedTask.itemId,
          minThreshold: savedTask.minThreshold,
          maxThreshold: savedTask.maxThreshold,
          isActive: !!savedTask.isActive,
          // 防止 recipeSnapshot 丢失
          recipeSnapshot: savedTask.recipeSnapshot || detailTask.value.recipeSnapshot
        })
      }
      if (typeof upsertTask === 'function') {
        upsertTask(detailTask.value)
      }
      ElMessage.success('阈值已更新')
    } catch (error) {
      if (error?.name === 'AbortError') {
        ElMessage.error('保存阈值超时，请稍后重试')
      } else {
        ElMessage.error('保存阈值失败')
      }
    } finally {
      clearTimeout(timeoutId)
      detailSaving.value = false
    }
  }

  async function deleteTask(itemId) {
    try {
      const response = await fetch(`${getApiBase()}/autocraft/tasks/${encodeURIComponent(itemId)}`, {
        method: 'DELETE'
      })
      if (!response.ok) {
        ElMessage.error('删除任务失败')
        return
      }

      autoCraftTasks.value = autoCraftTasks.value.filter((taskEntry) => taskEntry.itemId !== itemId)
      if (detailTask.value?.itemId === itemId) {
        detailVisible.value = false
        detailTask.value = null
      }
    } catch (_error) {
      ElMessage.error('删除任务失败')
    }
  }

  async function handleTaskActiveChange(task, isActive) {
    try {
      const response = await fetch(`${getApiBase()}/autocraft/tasks/${encodeURIComponent(task.itemId)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isActive })
      })

      if (!response.ok) {
        ElMessage.error('更新任务状态失败')
        return
      }

      const updatedTask = await response.json()
      upsertTask({
        itemId: updatedTask.itemId,
        itemName: updatedTask.itemName || updatedTask.itemId,
        minThreshold: updatedTask.minThreshold,
        maxThreshold: updatedTask.maxThreshold,
        isActive: !!updatedTask.isActive,
        recipeSnapshot: updatedTask.recipeSnapshot || task.recipeSnapshot
      })
    } catch (_error) {
      ElMessage.error('更新任务状态失败')
    }
  }

  return {
    factories,
    autoCraftTasks,
    wizardVisible,
    wizardStep,
    craftables,
    craftablesLoading,
    craftableQuery,
    selectedCraftable,
    minThreshold,
    maxThreshold,
    recipeTree,
    recipeLoading,
    detailVisible,
    detailTask,
    detailMinThreshold,
    detailMaxThreshold,
    detailSaving,
    systemStatus,
    inventoryIndex,
    taskIndex,
    energyPercent,
    energyColor,
    storagePercent,
    storageColor,
    storageTotalCapacity,
    storageTotalUsed,
    storageInternalRatio,
    storageExternalRatio,
    storageInternalUsage,
    storageExternalUsage,
    netRateClass,
    filteredCraftables,
    formatCompact,
    formatRate,
    formatTime,
    applyUpdatePayload,
    loadAutoCraftTasks,
    openWizard,
    closeWizard,
    selectCraftable,
    prevWizardStep,
    nextWizardStep,
    finishWizard,
    openTaskDetail,
    saveTaskThresholds,
    deleteTask,
    handleTaskActiveChange,
    displayItemName,
    fetchCraftables
  }
}
