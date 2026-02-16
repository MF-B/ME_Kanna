local M = {}

function M.inventoryUpdate(deviceId, name, isActive, rawItems, energy, storage, whitelistVersion)
    -- 直接构建，不搞那些 if energy then ... end 的重复劳动
    -- 因为 ae_bridge.lua 已经保证了 energy 和 storage 是完整的 table (即使值是0)
    local payload = {
        type = "update",
        id = deviceId,
        data = {
            name = name,
            active = isActive, -- 简化 key 为 active
            items = rawItems,  -- 简化 key 为 items (去掉 raw_)
            energy = energy,
            storage = storage
        },
    }
    
    if whitelistVersion then
        payload.whitelist_version = whitelistVersion
    end
    
    return payload
end

function M.productionFlow(factoryId, factoryName, delta, itemId)
    return {
        type = "production_flow",
        id = factoryId,
        name = factoryName,
        delta = delta,
        itemId = itemId
    }
end

function M.craftablesUpdate(deviceId, craftables, requestId)
    local payload = {
        type = "craftables",
        id = deviceId,
        craftables = craftables or {}
    }
    if requestId and requestId ~= "" then
        payload.requestId = requestId
    end
    return payload
end

function M.craftResult(deviceId, itemId, count, success, taskId, errMsg)
    return {
        type = "craft_result",
        id = deviceId,
        itemId = itemId,
        count = count or 0,
        success = success,
        taskId = taskId or "",
        error = errMsg or ""
    }
end

function M.patternsUpdate(deviceId, patterns, requestId)
    local payload = {
        type = "patterns",
        id = deviceId,
        patterns = patterns or {}
    }
    if requestId and requestId ~= "" then
        payload.requestId = requestId
    end
    return payload
end

function M.craftStatus(deviceId, taskId, isError, message)
    return {
        type = "craft_status",
        id = deviceId,
        taskId = tostring(taskId or ""),
        error = isError and true or false,
        message = message or ""
    }
end

return M
