import { reactive, shallowReactive } from 'vue'

const itemInfoCache = new Map()
const pending = new Map()
const sharedNames = reactive({})
const sharedIcons = reactive({})

function getApiBase() {
  const host = window.location.hostname
  const protocol = window.location.protocol
  return `${protocol}//${host}:8080`
}

export function useItemNames() {
  const toAbsoluteIconUrl = (icon) => {
    if (!icon) return ''
    if (/^https?:\/\//i.test(icon)) return icon
    return `${getApiBase()}${icon.startsWith('/') ? '' : '/'}${icon}`
  }

  const setItemInfo = (id, info, cache = true) => {
    if (!id) return

    const normalized = {
      name: info?.name || id,
      icon: toAbsoluteIconUrl(info?.icon)
    }

    if (cache) {
      itemInfoCache.set(id, normalized)
    }

    // 使用 reactive 增量更新，不再替换整个对象
    sharedNames[id] = normalized.name
    sharedIcons[id] = normalized.icon
  }

  const fetchItemInfo = async (id) => {
    if (!id || itemInfoCache.has(id) || pending.has(id)) return
    
    const url = `${getApiBase()}/api/item/${encodeURIComponent(id)}`
    const promise = fetch(url)
      .then((res) => {
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`)
        }
        return res.json()
      })
      .then((data) => {
        const hasUsefulData = !!(data?.name && data.name !== id) || !!data?.icon
        setItemInfo(id, data, hasUsefulData)
      })
      .catch(() => {
        // 失败也记入 cache，防止重复请求不存在的 ID
        setItemInfo(id, { name: id, icon: '' }, true)
      })
      .finally(() => {
        pending.delete(id)
      })

    pending.set(id, promise)
  }

  const ensureItemInfo = (id) => {
    if (!id) return
    
    // 如果 cache 里有，且响应式对象已经同步，则不再触发多余更新
    if (itemInfoCache.has(id)) {
      const cached = itemInfoCache.get(id)
      if (sharedNames[id] !== cached.name) {
        sharedNames[id] = cached.name
      }
      if (sharedIcons[id] !== cached.icon) {
        sharedIcons[id] = cached.icon
      }
      return
    }
    
    // 异步拉取
    fetchItemInfo(id)
  }

  const ensureName = (id) => {
    ensureItemInfo(id)
  }

  const ensureIcon = (id) => {
    ensureItemInfo(id)
  }

  return {
    names: sharedNames,
    icons: sharedIcons,
    ensureName,
    ensureIcon,
    ensureItemInfo
  }
}
