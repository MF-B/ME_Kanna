local aeBridge = require("ae_bridge")
local packets  = require("packets")
local util     = require("util")
local config   = require("config")
local whitelist = require("whitelist")

local M = {}

-- ========== 上下文 ==========
-- ctx 包含共享状态，由 main.lua 传入：
--   ctx.ws        : WebSocket 连接
--   ctx.getBridge : function() 返回当前 ae_device (可能为 nil)
--   ctx.setBridge : function(b) 设置 ae_device

-- ========== 各命令处理函数 ==========

local function handleConfigSync(packet, ctx)
    if whitelist.handlePacket(packet) then
        print("WS: Config Updated!")
        return true
    end
    return false
end

local function handleCmdCraftables(packet, ctx)
    local bridge = ctx.getBridge()
    if bridge then
        local craftableList = aeBridge.getCraftablesNormalized(bridge)
        util.sendJson(ctx.ws, packets.craftablesUpdate(config.DEVICE_ID, craftableList, packet.requestId))
        print("Craftables sent: " .. tostring(#craftableList))
    else
        print("Craftables request failed: no ME Bridge")
    end
end

local function handleCraft(packet, ctx)
    local itemId = packet.itemId
    local count = tonumber(packet.count) or 1
    if not itemId or itemId == "" then
        print("Craft ignored: missing item id")
        return
    end

    local bridge = ctx.getBridge()
    if not bridge then
        print("Craft failed: no ME Bridge")
        util.sendJson(ctx.ws, packets.craftResult(config.DEVICE_ID, itemId, count, false, nil, "no ME Bridge"))
        return
    end

    local task, err = aeBridge.craft(bridge, itemId, count)
    if task then
        local craftId = task.id or ""
        print("Craft queued: " .. tostring(itemId) .. " x" .. tostring(count) .. " (id=" .. tostring(craftId) .. ")")
        util.sendJson(ctx.ws, packets.craftResult(config.DEVICE_ID, itemId, count, true, craftId, nil))
    else
        print("Craft failed: " .. tostring(itemId) .. " x" .. tostring(count) .. " reason=" .. tostring(err))
        util.sendJson(ctx.ws, packets.craftResult(config.DEVICE_ID, itemId, count, false, nil, tostring(err)))
    end
end

-- ========== 命令分发表 ==========
-- key = packet.type, value = handler(packet, ctx)
-- config_sync 比较特殊：它没有固定的 type 字段，由 whitelist.handlePacket 判断

local handlers = {
    cmd_craftables = handleCmdCraftables,
    craft          = handleCraft,
}

-- ========== 公共接口 ==========

--- 分发一个 packet 到对应的 handler
--- @param packet table 解析后的 JSON 数据包
--- @param ctx table 上下文 (ws, getBridge, setBridge)
--- @return boolean 是否有 handler 处理了这个 packet
function M.dispatch(packet, ctx)
    if not packet then return false end

    -- 优先检查 config_sync (whitelist 有自己的判断逻辑)
    if handleConfigSync(packet, ctx) then
        return true
    end

    -- 按 type 分发
    local handler = handlers[packet.type]
    if handler then
        -- 需要 bridge 的命令：自动 ensureBridge
        ctx.setBridge(aeBridge.ensureBridge(ctx.getBridge()))
        handler(packet, ctx)
        return true
    end

    return false
end

--- 注册一个新的命令处理函数（供后续扩展用）
--- @param cmdType string 命令类型
--- @param handler function(packet, ctx) 处理函数
function M.register(cmdType, handler)
    handlers[cmdType] = handler
end

return M
