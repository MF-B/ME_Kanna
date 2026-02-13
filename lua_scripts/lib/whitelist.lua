local M = {
    lookup = {},
    list = {},
    version = 0,
    nextSyncAt = 0
}

local function nowMs()
    if os.epoch then
        return os.epoch("utc")
    end
    return math.floor(os.clock() * 1000)
end

local function normalizeList(data)
    local list = {}
    if data == nil then
        return list
    end
    if type(data) == "table" and data.monitored_items then
        data = data.monitored_items
    end
    if type(data) ~= "table" then
        return list
    end
    for _, item in ipairs(data) do
        if type(item) == "string" and item ~= "" then
            list[#list + 1] = item
        end
    end
    table.sort(list)
    return list
end

local function computeVersion(list)
    local hash = 5381
    for _, item in ipairs(list) do
        for i = 1, #item do
            hash = (hash * 33 + string.byte(item, i)) % 2147483647
        end
        hash = (hash * 33 + 10) % 2147483647
    end
    return hash
end

function M.update(data, new_version)
    local list = normalizeList(data)
    local version = new_version
    if version == nil then
        version = computeVersion(list)
    end

    if M.version == version and #M.list == #list then
        return false
    end

    M.list = list
    M.lookup = {}
    for _, item in ipairs(M.list) do
        M.lookup[item] = true
    end
    M.version = version
    print("Config Updated to Version: " .. tostring(version))
    return true
end

function M.getList()
    return M.list
end

function M.getVersion()
    return M.version
end

function M.sync(apiUrl, intervalSeconds, force)
    if not apiUrl or apiUrl == "" then
        return false
    end

    local intervalMs = (intervalSeconds or 0) * 1000
    local now = nowMs()
    if not force and intervalMs > 0 and M.nextSyncAt > 0 and now < M.nextSyncAt then
        return false
    end
    if intervalMs > 0 then
        M.nextSyncAt = now + intervalMs
    end

    local resp = http.get(apiUrl)
    if not resp then
        return false
    end

    local body = resp.readAll()
    resp.close()

    local ok, decoded = pcall(textutils.unserializeJSON, body)
    if not ok or type(decoded) ~= "table" then
        return false
    end

    return M.update(decoded.monitored_items or decoded.data or decoded, decoded.version)
end

function M.handlePacket(packet)
    if not packet or type(packet) ~= "table" then
        return false
    end
    if packet.type == "config_sync" then
        return M.update(packet.data or packet.monitored_items or packet, packet.version)
    end
    return false
end

return M