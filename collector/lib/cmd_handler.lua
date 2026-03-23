local aeBridge = require("ae_bridge")

local M = {}

local function handleCraftItem(packet, ctx)
    local data = packet.data or {}
    local itemId = data.item
    local count = tonumber(data.amount) or 1

    if type(itemId) ~= "string" or itemId == "" then
        print("cmd_craft_item ignored: invalid item")
        return
    end

    local bridge = aeBridge.ensureBridge(ctx.getBridge())
    ctx.setBridge(bridge)

    if not bridge then
        print("cmd_craft_item failed: no ME Bridge")
        return
    end

    local task, err = aeBridge.craft(bridge, itemId, count)
    if task then
        local craftId = task.id or ""
        print("Craft queued: " .. tostring(itemId) .. " x" .. tostring(count) .. " (id=" .. tostring(craftId) .. ")")
    else
        print("Craft failed: " .. tostring(itemId) .. " x" .. tostring(count) .. " reason=" .. tostring(err))
    end
end

local function handleSetPriority(packet)
    local data = packet.data or {}
    aeBridge.updatePriorityWatchlist(data.items)
    local count = type(data.items) == "table" and #data.items or 0
    print("Priority watchlist updated: " .. tostring(count))
end

local function handleScanCraftables(packet, ctx)
    local bridge = aeBridge.ensureBridge(ctx.getBridge())
    ctx.setBridge(bridge)

    if not bridge then
        print("scan ignored: no ME Bridge")
        return
    end

    print("Scanning craftables...")
    local items = aeBridge.scanCraftables(bridge)
    local payload = require("packets").evtCraftables(items)
    require("util").sendJson(ctx.ws, payload)
    print("Scan complete: " .. tostring(#items) .. " items")
end

local handlers = {
    cmd_craft_item = handleCraftItem,
    cmd_set_priority = handleSetPriority,
    cmd_set_routine = handleSetRoutine,
    cmd_scan_craftables = handleScanCraftables,
}

function M.dispatch(packet, ctx)
    if type(packet) ~= "table" or type(packet.type) ~= "string" then
        return false
    end

    local handler = handlers[packet.type]
    if not handler then
        return false
    end

    handler(packet, ctx)
    return true
end

function M.register(cmdType, handler)
    handlers[cmdType] = handler
end

return M
