local M = {}

local priorityList = {}
local routineList = {}
local routineCursor = 1
local ROUTINE_BATCH_SIZE = 64

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
        stored = bridge.getStoredEnergy() or 0,
        capacity = bridge.getEnergyCapacity() or 0,
        usage = bridge.getEnergyUsage() or 0,
        input = bridge.getAverageEnergyInput() or 0,
    }
end

function M.collectStorage(bridge)
    return {
        itemTotal = bridge.getTotalItemStorage() or 0,
        itemUsed = bridge.getUsedItemStorage() or 0,
        fluidTotal = bridge.getTotalFluidStorage() or 0,
        fluidUsed = bridge.getUsedFluidStorage() or 0,
    }
end

local function normalizeItems(items)
    local result = {}
    if type(items) ~= "table" then
        return result
    end

    for _, itemId in ipairs(items) do
        if type(itemId) == "string" and itemId ~= "" then
            table.insert(result, itemId)
        end
    end

    return result
end

function M.updatePriorityWatchlist(items)
    priorityList = normalizeItems(items)
end

function M.updateRoutineWatchlist(items)
    routineList = normalizeItems(items)
    if routineCursor > #routineList then
        routineCursor = 1
    end
end

local function appendItemCount(result, seen, bridge, itemId)
    if seen[itemId] then
        return
    end

    seen[itemId] = true
    local detail = bridge.getItem({name = itemId})
    table.insert(result, {
        name = itemId,
        count = (detail and detail.count) or 0,
    })
end

function M.collectItems(bridge)
    local result = {}
    local seen = {}

    for _, itemId in ipairs(priorityList) do
        appendItemCount(result, seen, bridge, itemId)
    end

    local totalRoutine = #routineList
    if totalRoutine > 0 then
        local startIndex = routineCursor
        local count = math.min(ROUTINE_BATCH_SIZE, totalRoutine)

        for offset = 0, count - 1 do
            local idx = ((startIndex + offset - 1) % totalRoutine) + 1
            appendItemCount(result, seen, bridge, routineList[idx])
        end

        routineCursor = ((startIndex + count - 1) % totalRoutine) + 1
    end

    return result
end

function M.collectCPUs(bridge)
    local raw = bridge.getCraftingCPUs() or {}
    local result = {}

    if type(raw) ~= "table" then
        return result
    end

    for _, cpu in ipairs(raw) do
        local job = cpu and cpu.craftingJob
        local normalizedJob = nil

        if type(job) == "table" and type(job.name) == "string" and job.name ~= "" then
            normalizedJob = {
                name = job.name,
                count = tonumber(job.count) or 0,
            }
        end

        table.insert(result, {
            coProcessors = tonumber(cpu and cpu.coProcessors) or 0,
            storage = tonumber(cpu and cpu.storage) or 0,
            craftingJob = normalizedJob,
        })
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

function M.scanCraftables(bridge)
    local result = {}
    if not bridge or not bridge.getCraftableItems then
        return result
    end

    local items = bridge.getCraftableItems()
    if type(items) == "table" then
        for _, item in ipairs(items) do
            if type(item) == "table" and type(item.name) == "string" and item.name ~= "" then
                table.insert(result, {
                    name = item.name,
                    count = tonumber(item.count) or 0
                })
            end
        end
    end
    return result
end

return M
