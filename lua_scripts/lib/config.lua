local M = {}

-- ================= 配置区域 =================
M.HOST = "http://192.168.1.34:8080"
M.WS_URL = "ws://192.168.1.34:8080/ws/minecraft"
M.API_URL = M.HOST .. "/config/whitelist"
M.DEVICE_ID = "ae_hub"

M.FACTORY_ID = ""
M.SIDE_REDSTONE = "top"

M.RECONNECT_DELAY = 5
M.SYNC_INTERVAL = 60
-- ===========================================

return M
