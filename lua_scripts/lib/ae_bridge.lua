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

-- 参数说明：
-- bridge: 外设对象
-- monitorList: 这是一个纯 ID 列表，例如 {"minecraft:iron_ingot", "ae2:silicon"}
function M.collectFilteredItems(bridge, monitorList)
    local result = {}

    for _, id in ipairs(monitorList) do
        local item = bridge.getItem({name = id})

        if item then
            result[id] = item.count or 0
        else
            result[id] = 0
        end
    end
    return result
end

return M
