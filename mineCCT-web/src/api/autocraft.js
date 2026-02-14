import { requestJson } from './http'

export async function fetchCraftables({ signal } = {}) {
  const { res, data } = await requestJson('/autocraft/craftables', { signal })
  return { res, data }
}

export async function fetchTasks() {
  const { res, data } = await requestJson('/autocraft/tasks')
  return { res, data }
}

export async function fetchRecipe({ itemId }) {
  const { res, data } = await requestJson(`/autocraft/recipe?itemId=${encodeURIComponent(itemId)}`)
  return { res, data }
}

export async function createTask(payload, { signal } = {}) {
  const { res, data } = await requestJson('/autocraft/tasks', {
    method: 'POST',
    body: payload,
    signal
  })
  return { res, data }
}

export async function deleteTask({ itemId }) {
  const { res, data } = await requestJson(`/autocraft/tasks/${encodeURIComponent(itemId)}`, {
    method: 'DELETE'
  })
  return { res, data }
}

export async function updateTaskActive({ itemId, isActive }) {
  const { res, data } = await requestJson(`/autocraft/tasks/${encodeURIComponent(itemId)}`, {
    method: 'PATCH',
    body: { isActive }
  })
  return { res, data }
}
