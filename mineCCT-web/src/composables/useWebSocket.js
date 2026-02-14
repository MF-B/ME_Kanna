import { ref, onUnmounted } from 'vue'

export function useWebSocket({ onUpdate, onOpen, onClose, reconnectDelay = 3000 } = {}) {
  const connected = ref(false)
  let socket = null
  let reconnectTimer = null

  function getWsUrl() {
    let host = window.location.hostname
    if (host.includes(':') && !host.startsWith('[')) {
      host = `[${host}]`
    }
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    return `${protocol}://${host}:8080/ws/web`
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  function scheduleReconnect() {
    clearReconnectTimer()
    reconnectTimer = setTimeout(connect, reconnectDelay)
  }

  function connect() {
    clearReconnectTimer()

    if (socket) {
      socket.onopen = null
      socket.onclose = null
      socket.onerror = null
      socket.onmessage = null
      socket.close()
    }

    socket = new WebSocket(getWsUrl())

    socket.onopen = () => {
      connected.value = true
      if (onOpen) onOpen()
    }

    socket.onclose = () => {
      connected.value = false
      if (onClose) onClose()
      scheduleReconnect()
    }

    socket.onerror = () => {
      if (socket) socket.close()
    }

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data)
        if (onUpdate) onUpdate(payload)
      } catch (_error) {
      }
    }
  }

  function send(payload) {
    if (!socket || socket.readyState !== WebSocket.OPEN) return false
    socket.send(JSON.stringify(payload))
    return true
  }

  function disconnect() {
    clearReconnectTimer()
    if (socket) {
      socket.close()
      socket = null
    }
  }

  onUnmounted(disconnect)

  return {
    connected,
    connect,
    disconnect,
    send
  }
}
