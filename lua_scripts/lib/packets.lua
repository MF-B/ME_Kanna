local M = {}

function M.inventoryUpdate(deviceId, name, isActive, rawItems)
    return {
        type = "update",
        id = deviceId,
        data = {
            [deviceId] = {
                name = name,
                isActive = isActive,
                raw_items = rawItems
            }
        }
    }
end

function M.productionFlow(factoryId, delta, itemName)
    return {
        type = "production_flow",
        id = factoryId,
        delta = delta,
        item = itemName
    }
end

return M
