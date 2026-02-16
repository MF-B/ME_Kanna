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

function M.getCraftablesNormalized(bridge, filter)
    local raw = M.getCraftables(bridge, filter)
    local result = {}
    if type(raw) ~= "table" then return result end

    for _, entry in ipairs(raw) do
        local itemId = entry and entry.name
        if type(itemId) == "string" and itemId ~= "" then
            table.insert(result, {
                itemId = itemId,
                itemName = entry.displayName or itemId,
                fingerprint = entry.fingerprint,
                count = entry.count or 0
            })
        end
    end

    return result
end

function M.craft(bridge, itemId, count)
    if not bridge or not bridge.craftItem then return nil, "no craftItem api" end
    if type(itemId) ~= "string" or itemId == "" then return nil, "invalid itemId" end

    local craftCount = tonumber(count) or 1
    if craftCount < 1 then craftCount = 1 end

    local filter = {name = itemId, count = craftCount}
    if bridge.isCraftable and not bridge.isCraftable({name = itemId}) then
        return nil, "item is not craftable"
    end

    local task, err = bridge.craftItem(filter)
    if task then
        return task, nil
    end

    return nil, err or "craft failed"
end

-- ========== 合成任务查询 ==========

function M.getCraftingTasks(bridge)
    if not bridge or not bridge.getCraftingTasks then return {} end
    return bridge.getCraftingTasks() or {}
end

function M.getCraftingTask(bridge, taskId)
    if not bridge or not bridge.getCraftingTask then return nil, "no getCraftingTask api" end
    return bridge.getCraftingTask(taskId)
end

function M.getCraftingCPUs(bridge)
    if not bridge or not bridge.getCraftingCPUs then return {} end
    return bridge.getCraftingCPUs() or {}
end

-- ========== 配方查询 ==========

function M.getPatterns(bridge, filter)
    if not bridge or not bridge.getPatterns then return {} end
    return bridge.getPatterns(filter or {}) or {}
end

-- ========== 任务控制 ==========

function M.cancelCraftingTasks(bridge, filter)
    if not bridge or not bridge.cancelCraftingTasks then return 0 end
    return bridge.cancelCraftingTasks(filter or {}) or 0
end

function M.isCraftingItem(bridge, filter)
    if not bridge or not bridge.isCrafting then return false end
    return bridge.isCrafting(filter or {})
end

return M
