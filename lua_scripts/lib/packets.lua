local M = {}

function M.inventoryUpdate(deviceId, name, isActive, rawItems, energy, storage)
    local report = {
        name = name,
        isActive = isActive,
        raw_items = rawItems
    }

    if energy then
        report.energy = {
            energyStored = energy.energyStored,
            energyMax = energy.energyMax,
            energyUsage = energy.energyUsage,
            averageEnergyInput = energy.averageEnergyInput
        }
    end

    if storage then
        report.storage = storage
    end

    return {
        type = "update",
        id = deviceId,
        data = {
            [deviceId] = report
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
