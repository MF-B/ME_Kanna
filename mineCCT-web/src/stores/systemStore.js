import { defineStore } from 'pinia'
import { computed, ref, shallowRef } from 'vue'
import * as autoCraftApi from '../api/autocraft'
import { useItemNames } from '../composables/useItemNames'

export const useSystemStore = defineStore('system', () => {
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

  const inventoryIndex = shallowRef({})
  const itemRates = shallowRef({})

  const { names: itemNameMap, ensureName } = useItemNames()

  const defaultMinThreshold = 64
  const defaultMaxThreshold = 256

  function resetFactories() {
    factories.value = []
  }

  function setWizardVisible(value) {
    wizardVisible.value = !!value
  }

  function setCraftableQuery(value) {
    craftableQuery.value = value || ''
  }

  function setMinThreshold(value) {
    const v = Number(value)
    if (Number.isFinite(v)) minThreshold.value = v
  }

  function setMaxThreshold(value) {
    const v = Number(value)
    if (Number.isFinite(v)) maxThreshold.value = v
  }

  function setDetailVisible(value) {
    detailVisible.value = !!value
    if (!detailVisible.value) {
      detailTask.value = null
      detailSaving.value = false
    }
  }

  function setDetailMinThreshold(value) {
    const v = Number(value)
    if (Number.isFinite(v)) detailMinThreshold.value = v
  }

  function setDetailMaxThreshold(value) {
    const v = Number(value)
    if (Number.isFinite(v)) detailMaxThreshold.value = v
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

  function applyUpdatePayload(payload) {
    if (payload.type === 'craftables') {
      const list = payload.craftables || []
      craftables.value = list.map((c) => ({
        itemId: c.itemId,
        itemName: c.itemName || c.itemId,
        count: c.count || 0
      })).filter((c) => c.itemId)
      craftables.value.forEach((c) => ensureName(c.itemId))
      if (craftablesLoading.value) craftablesLoading.value = false
      return
    }

    if (payload.type !== 'update') return

    // 1. 更新基础数据
    factories.value = payload.data || []
    if (payload.system) {
      systemStatus.value = payload.system
    }

    // 2. 批量预取 itemId 名称（去重）
    const itemIdsToFetch = new Set()

    // 从工厂产物中收集
    factories.value.forEach(f => {
      if (f.items) {
        Object.keys(f.items).forEach(id => itemIdsToFetch.add(id))
      }
    })

    // 从系统库存中收集
    if (payload.system?.inventory) {
      Object.keys(payload.system.inventory).forEach(id => itemIdsToFetch.add(id))
    }

    // 执行预取
    itemIdsToFetch.forEach(id => ensureName(id))

    // 3. 优化 inventoryIndex：优先使用 system.inventory，否则从 factories 计算
    const systemInventory = payload.system?.inventory
    if (systemInventory && Object.keys(systemInventory).length > 0) {
      inventoryIndex.value = systemInventory
      // fallthrough to calc rates
    } else {
      const nextIndex = {}
      factories.value.forEach((factoryData) => {
        const factoryItems = factoryData?.items || {}
        Object.values(factoryItems).forEach((factoryItem) => {
          if (factoryItem?.itemId) {
            nextIndex[factoryItem.itemId] = factoryItem.count || 0
          }
        })
      })
      // Simple diff check omitted for brevity as we are rewriting this block
      inventoryIndex.value = nextIndex
    }

    // 4. Calculate Item Rates (Sum of prodRate from all factories)
    const nextRates = {}
    factories.value.forEach((factoryData) => {
      const factoryItems = factoryData?.items || {}
      Object.values(factoryItems).forEach((factoryItem) => {
        if (factoryItem?.itemId && factoryItem.prodRate) {
          nextRates[factoryItem.itemId] = (nextRates[factoryItem.itemId] || 0) + factoryItem.prodRate
        }
      })
    })
    itemRates.value = nextRates
  }

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

  const storagePercent = computed(() => {
    if (!storageTotalCapacity.value) return 0
    const p = (storageTotalUsed.value / storageTotalCapacity.value) * 100
    return Math.min(Math.max(p, 0), 100)
  })

  const storageColor = computed(() => storagePercent.value > 90 ? '#ff6b6b' : '#5d8aff')

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

  function displayItemName(itemId, fallbackName) {
    if (!itemId) return fallbackName || ''
    return itemNameMap[itemId] || fallbackName || itemId
  }

  const filteredCraftables = computed(() => {
    const query = craftableQuery.value.trim().toLowerCase()
    if (!query) return craftables.value
    return craftables.value.filter((craftableItem) => {
      const itemIdKeyword = (craftableItem.itemId || '').toLowerCase()
      const itemNameKeyword = displayItemName(craftableItem.itemId, craftableItem.itemName).toLowerCase()
      return itemIdKeyword.includes(query) || itemNameKeyword.includes(query)
    })
  })

  async function fetchCraftables() {
    if (craftablesLoading.value) return
    craftablesLoading.value = true
    const abortController = new AbortController()
    const timeoutId = setTimeout(() => abortController.abort(), 8000)
    try {
      const { res, data } = await autoCraftApi.fetchCraftables({ signal: abortController.signal })
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }
      const craftableList = Array.isArray(data) ? data : (data.items || [])
      craftables.value = craftableList.map((craftableEntry) => {
        if (typeof craftableEntry === 'string') {
          return { itemId: craftableEntry, itemName: craftableEntry }
        }
        return {
          itemId: craftableEntry.itemId,
          itemName: craftableEntry.itemName || craftableEntry.itemId,
          count: craftableEntry.count || 0
        }
      }).filter((craftableEntry) => craftableEntry.itemId)

      // 在批量预取中已经处理了数据流入，这里不再分散触发 ensureName
    } catch (err) {
      craftables.value = []
      if (err?.name === 'AbortError') {
        throw new Error('获取可合成列表超时')
      }
      throw new Error('获取可合成列表失败')
    } finally {
      clearTimeout(timeoutId)
      craftablesLoading.value = false
    }
  }

  async function loadAutoCraftTasks() {
    try {
      const { res, data } = await autoCraftApi.fetchTasks()
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
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
    } catch (err) {
      autoCraftTasks.value = []
      throw new Error(`加载自动合成任务失败: ${err.message || ''}`)
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

    // 触发首次加载，但不要在这里抛出未处理异常（UI 上也提供了手动“刷新”按钮）
    fetchCraftables().catch(() => { })
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
      const { res: recipeRes, data: recipeData } = await autoCraftApi.fetchRecipe({ itemId: main.itemId })
      if (!recipeRes.ok) throw new Error(`HTTP ${recipeRes.status}`)
      recipeTree.value = recipeData

      const taskPayload = {
        itemId: main.itemId,
        itemName: displayItemName(main.itemId, main.itemName),
        minThreshold: minThreshold.value,
        maxThreshold: maxThreshold.value,
        isActive: false,
        recipeSnapshot: recipeData
      }

      const { res: createRes, data: savedTask } = await autoCraftApi.createTask(taskPayload)

      if (!createRes.ok) {
        throw new Error('创建任务失败')
      }
      if (!savedTask) {
        throw new Error('创建任务失败：返回数据为空')
      }
      upsertTask({
        itemId: savedTask.itemId,
        itemName: savedTask.itemName || savedTask.itemId,
        minThreshold: savedTask.minThreshold,
        maxThreshold: savedTask.maxThreshold,
        isActive: !!savedTask.isActive,
        recipeSnapshot: savedTask.recipeSnapshot || recipeData
      })

      wizardVisible.value = false
    } catch (err) {
      throw new Error(err.message || '创建任务失败')
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
      throw new Error('阈值不合法：目标阈值必须大于等于最低阈值')
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

      const { res, data: savedTask } = await autoCraftApi.createTask(payload, {
        signal: abortController.signal
      })

      if (!res.ok) {
        const msg = (savedTask && savedTask.error) || '保存阈值失败'
        throw new Error(msg)
      }

      if (detailTask.value && savedTask) {
        Object.assign(detailTask.value, {
          itemId: savedTask.itemId,
          itemName: savedTask.itemName || savedTask.itemId,
          minThreshold: savedTask.minThreshold,
          maxThreshold: savedTask.maxThreshold,
          isActive: !!savedTask.isActive,
          recipeSnapshot: savedTask.recipeSnapshot || detailTask.value.recipeSnapshot
        })
      }

      upsertTask(detailTask.value)
    } catch (err) {
      if (err?.name === 'AbortError') {
        throw new Error('保存阈值超时，请稍后重试')
      }
      throw new Error(err.message || '保存阈值失败')
    } finally {
      clearTimeout(timeoutId)
      detailSaving.value = false
    }
  }

  async function deleteTask(itemId) {
    try {
      const { res } = await autoCraftApi.deleteTask({ itemId })
      if (!res.ok) {
        throw new Error('删除任务失败')
      }

      autoCraftTasks.value = autoCraftTasks.value.filter((taskEntry) => taskEntry.itemId !== itemId)
      if (detailTask.value?.itemId === itemId) {
        detailVisible.value = false
        detailTask.value = null
      }
    } catch (err) {
      throw new Error(err.message || '删除任务失败')
    }
  }

  async function handleTaskActiveChange(task, isActive) {
    try {
      const { res: response, data: updatedTask } = await autoCraftApi.updateTaskActive({
        itemId: task.itemId,
        isActive
      })

      if (!response.ok) {
        throw new Error('更新任务状态失败')
      }
      if (!updatedTask) {
        throw new Error('更新任务状态失败：返回数据为空')
      }
      upsertTask({
        itemId: updatedTask.itemId,
        itemName: updatedTask.itemName || updatedTask.itemId,
        minThreshold: updatedTask.minThreshold,
        maxThreshold: updatedTask.maxThreshold,
        isActive: !!updatedTask.isActive,
        recipeSnapshot: updatedTask.recipeSnapshot || task.recipeSnapshot
      })
    } catch (err) {
      throw new Error(err.message || '更新任务状态失败')
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
    itemRates,
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
    resetFactories,
    setWizardVisible,
    setCraftableQuery,
    setMinThreshold,
    setMaxThreshold,
    setDetailVisible,
    setDetailMinThreshold,
    setDetailMaxThreshold,
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
})
