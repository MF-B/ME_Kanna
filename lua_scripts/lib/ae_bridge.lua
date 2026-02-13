local M = {}

function M.findBridge()
    print("Scanning for ME Bridge...")
    local bridge = peripheral.find("me_bridge")
    if bridge then
        print(" + Linked: " .. peripheral.getName(bridge))
        return bridge
    end
    return nil
end

function M.ensureBridge(existing)
    if existing then return existing end
    return M.findBridge()
end

function M.collectEnergy(bridge)
    return {
        energyStored = bridge.getStoredEnergy() or 0,
        energyMax = bridge.getEnergyCapacity() or 0,
        energyUsage = bridge.getEnergyUsage() or 0,
        averageEnergyInput = bridge.getAverageEnergyInput() or 0
    }
end

function M.collectStorage(bridge)
    return {
        itemTotal = bridge.getTotalItemStorage() or 0,
        itemUsed = bridge.getUsedItemStorage() or 0,

        itemExternalTotal = bridge.getTotalExternalItemStorage() or 0,
        itemExternalUsed = bridge.getUsedExternalItemStorage() or 0,

        fluidTotal = bridge.getTotalFluidStorage() or 0,
        fluidUsed = bridge.getUsedFluidStorage() or 0,
    }
end

-- 参数说明：
-- bridge: 外设对象
-- monitorList: 这是一个纯 ID 列表，例如 {"minecraft:iron_ingot", "ae2:silicon"}
function M.collectFilteredItems(bridge, monitorList)
    local result = {}
    for _, id in ipairs(monitorList) do
        local itemDetail = bridge.getItem({name = id})
        -- 简化写法：如果有 item 取 count，没有取 0
        result[id] = (itemDetail and itemDetail.count) or 0
    end
    return result
end

function M.getCraftables(bridge, filter)
    if not bridge or not bridge.getCraftableItems then return {} end
    return bridge.getCraftableItems(filter or {}) or {}
end

function M.craft(bridge, itemId, count)
    if not bridge or not bridge.craftItem then return false end
    local filter = {name = itemId, count = count}
    return bridge.craftItem(filter)
end

return M
