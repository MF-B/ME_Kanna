local M = {}

function M.findBridge()
    print("Scanning for ME Bridge...")
    for _, name in pairs(peripheral.getNames()) do
        if peripheral.getType(name) == "me_bridge" then
            print(" + Linked: " .. name)
            return peripheral.wrap(name)
        end
    end
    return nil
end

function M.ensureBridge(existing)
    if existing then return existing end
    local bridge = M.findBridge()
    if not bridge then sleep(5) end
    return bridge
end

function M.collectFilteredItems(bridge, isMonitored)
    local filtered = {}
    local allItems = bridge.getItems()

    for _, item in pairs(allItems) do
        local count = item.count or 0
        local name = item.name
        if name and isMonitored(name) then
            filtered[name] = (filtered[name] or 0) + count
        end
    end

    return filtered
end

return M
