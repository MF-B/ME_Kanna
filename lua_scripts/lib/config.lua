local M = {}

-- ================= 配置区域 =================
M.HOST = "http://127.0.0.1:8080"
M.WS_URL = "ws://127.0.0.1:8080/ws/minecraft"
M.API_URL = M.HOST .. "/config/whitelist"
M.DEVICE_ID = "ae_hub"

M.FACTORY_ID = ""
M.SIDE_REDSTONE = "top"

M.RECONNECT_DELAY = 5
M.SYNC_INTERVAL = 60
-- ===========================================

return M
