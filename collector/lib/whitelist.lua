local util = require("util") -- 引入 util
local M = {
    lookup = {},
    list = {},
    version = "0", -- 默认为字符串
    nextSyncAt = 0
}

-- 只保留基础的清洗逻辑
local function normalizeList(data)
    local list = {}
    if not data then return list end
    
    -- 兼容 {monitored_items = [...]} 和纯 [...]
    local items = (type(data) == "table" and data.monitored_items) or data
    if type(items) ~= "table" then return items end -- 如果已经是 table 且没 key，可能就是纯数组
    
    -- 简单的过滤
    for _, item in ipairs(items) do
        if type(item) == "string" and item ~= "" then
            table.insert(list, item)
        end
    end
    return list
end

function M.update(data, new_version)
    -- 如果后端没给版本号，用时间戳兜底
    local version = new_version or tostring(os.time())
    
    -- 简单的字符串比对
    if M.version == version then
        return false
    end

    M.list = normalizeList(data)
    M.lookup = {}
    for _, item in ipairs(M.list) do
        M.lookup[item] = true
    end
    M.version = version
    print("Config Updated: " .. tostring(version))
    return true
end

function M.getList() return M.list end
function M.getVersion() return M.version end

function M.sync(apiUrl, intervalSeconds, force)
    if not apiUrl or apiUrl == "" then return false end

    local intervalMs = (intervalSeconds or 0) * 1000
    local now = util.nowMs() -- 使用 util
    
    if not force and intervalMs > 0 and M.nextSyncAt > 0 and now < M.nextSyncAt then
        return false
    end
    
    if intervalMs > 0 then M.nextSyncAt = now + intervalMs end

    local resp = http.get(apiUrl)
    if not resp then return false end

    local body = resp.readAll()
    resp.close()

    local ok, decoded = pcall(textutils.unserializeJSON, body)
    if not ok or type(decoded) ~= "table" then return false end

    return M.update(decoded.monitored_items or decoded, decoded.version)
end

function M.handlePacket(packet)
    if packet and packet.type == "config_sync" then
        return M.update(packet.data, packet.version)
    end
    return false
end

return M