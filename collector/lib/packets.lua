local M = {}

local function envelope(msgType, data)
    return {
        type = msgType,
        data = data or {},
    }
end

function M.evtTick(items, cpus, energy, storage)
    return envelope("evt_tick", {
        items = items or {},
        cpus = cpus or {},
        energy = energy or {},
        storage = storage or {},
    })
end

function M.evtCrafting(jobId, isError, message)
    return envelope("evt_crafting", {
        jobId = tonumber(jobId) or 0,
        isError = isError and true or false,
        message = message or "",
    })
end

function M.evtProduction(factoryId, factoryName, itemId, delta)
    return envelope("evt_production", {
        factoryId = factoryId or "",
        factoryName = factoryName or "",
        itemId = itemId or "",
function M.evtCraftables(items)
    return envelope("evt_craftables", {
        items = items or {},
    })
end

return M
