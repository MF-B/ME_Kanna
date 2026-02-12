import { ref } from 'vue'

const nameCache = new Map()
const pending = new Map()
const sharedNames = ref({})

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
    sharedNames.value = { ...sharedNames.value, [id]: name }
  }

  const fetchName = async (id) => {
    if (!id || nameCache.has(id) || pending.has(id)) return
    const url = `${getApiBase()}/item-name/${encodeURIComponent(id)}`
    const request = fetch(url)
      .then((res) => res.json())
      .then((data) => {
        const name = data?.name || id
        const shouldCache = name && name !== id
        setName(id, name, shouldCache)
      })
      .catch(() => {
        setName(id, id, false)
      })
      .finally(() => {
        pending.delete(id)
      })

    pending.set(id, request)
  }

  const ensureName = (id) => {
    if (!id) return
    if (nameCache.has(id)) {
      setName(id, nameCache.get(id))
      return
    }
    const current = sharedNames.value[id]
    if (current === id) {
      fetchName(id)
      return
    }
    fetchName(id)
  }

  return {
    names: sharedNames,
    ensureName
  }
}
