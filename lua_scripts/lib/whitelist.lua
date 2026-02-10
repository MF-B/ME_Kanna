local M = {
    items = {},
    lastSync = 0
}

function M.sync(apiUrl)
    print("Syncing config...")
    local response = http.get(apiUrl)
    if response then
        local body = response.readAll()
        response.close()
        local data = textutils.unserializeJSON(body)
        if data and data.monitored_items then
            M.items = {}
            for _, item in pairs(data.monitored_items) do
                M.items[item] = true
            end
            M.lastSync = os.clock()
            print("Config Synced! Count: " .. tostring(#data.monitored_items))
            return true
        end
    end
    print("Config Sync Failed!")
    return false
end

function M.isMonitored(itemId)
    return M.items[itemId] == true
end

return M
