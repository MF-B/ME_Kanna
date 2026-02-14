import { reactive, shallowReactive } from 'vue'

const nameCache = new Map()
const pending = new Map()
const sharedNames = reactive({})

function getApiBase() {
  const host = window.location.hostname
  const protocol = window.location.protocol
  return `${protocol}//${host}:8080`
}

export function useItemNames() {
  const setName = (id, name, cache = true) => {
    if (!id) return
    if (cache) {
      nameCache.set(id, name)
    }
    // 使用 reactive 增量更新，不再替换整个 sharedNames.value
    sharedNames[id] = name
  }

  const fetchName = async (id) => {
    if (!id || nameCache.has(id) || pending.has(id)) return
    
    const url = `${getApiBase()}/item-name/${encodeURIComponent(id)}`
    const promise = fetch(url)
      .then((res) => res.json())
      .then((data) => {
        const name = data?.name || id
        const shouldCache = name && name !== id
        setName(id, name, shouldCache)
      })
      .catch(() => {
        // 失败也记入 cache，防止重复请求不存在的 ID
        setName(id, id, true)
      })
      .finally(() => {
        pending.delete(id)
      })

    pending.set(id, promise)
  }

  const ensureName = (id) => {
    if (!id) return
    
    // 如果 cache 里有，且 sharedNames 已经对应上，则什么都不做（不再触发 setName 引起响应式系统抖动）
    if (nameCache.has(id)) {
      if (sharedNames[id] !== nameCache.get(id)) {
        sharedNames[id] = nameCache.get(id)
      }
      return
    }
    
    // 异步拉取
    fetchName(id)
  }

  return {
    names: sharedNames,
    ensureName
  }
}
