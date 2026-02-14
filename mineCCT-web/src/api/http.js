function getApiBase() {
  const host = window.location.hostname
  const protocol = window.location.protocol
  return `${protocol}//${host}:8080`
}

export async function request(path, { method = 'GET', body, signal, headers } = {}) {
  const res = await fetch(`${getApiBase()}${path}`, {
    method,
    headers: {
      ...(body ? { 'Content-Type': 'application/json' } : null),
      ...(headers || null)
    },
    body: body ? JSON.stringify(body) : undefined,
    signal
  })

  return res
}

export async function requestJson(path, options) {
  const res = await request(path, options)
  const text = await res.text()
  let data = null
  try {
    data = text ? JSON.parse(text) : null
  } catch (_e) {
    data = null
  }

  return { res, data }
}
