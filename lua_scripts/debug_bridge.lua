package.path = "lib/?.lua;" .. package.path

local config = require("config")
local wsClient = require("ws_client")
local aeBridge = require("ae_bridge")
local packets = require("packets")
local util = require("util")

local DEBUG_ITEM_ID = "minecraft:iron_ingot"
local DEBUG_TASK_ID = nil
local SEND_INTERVAL_SECONDS = 30
local MAX_DEPTH = 6

local function sanitizeValue(value, visited, depth)
    if type(value) ~= "table" then
        return value
    end
    if visited[value] then
        return "<ref>"
    end
    if depth > MAX_DEPTH then
        return "<max_depth>"
    end

    visited[value] = true

    local out = {}
    local isArray = true
    local maxIndex = 0
    for k, _ in pairs(value) do
        if type(k) ~= "number" then
            isArray = false
            break
        end
        if k > maxIndex then
            maxIndex = k
        end
    end

    if isArray then
        for i = 1, maxIndex do
            out[i] = sanitizeValue(value[i], visited, depth + 1)
        end
        return out
    end

    for k, v in pairs(value) do
        local outKey = k
        if type(k) ~= "string" and type(k) ~= "number" then
            outKey = tostring(k)
        end
        out[outKey] = sanitizeValue(v, visited, depth + 1)
    end

    return out
end

local function callBridge(bridge, fnName, ...)
    if not bridge or type(bridge[fnName]) ~= "function" then
        return { ok = false, error = "missing bridge api: " .. tostring(fnName) }
    end

    local ok, result, err = pcall(bridge[fnName], ...)
    if not ok then
        return { ok = false, error = tostring(result) }
    end

    if result == nil and err ~= nil then
        return { ok = false, error = tostring(err) }
    end

    if result == nil then
        return { ok = true, dataType = "nil", dataValue = "nil" }
    end

    if type(result) ~= "table" then
        return { ok = true, dataType = type(result), dataValue = result }
    end

    return { ok = true, dataType = "table", data = sanitizeValue(result, {}, 0) }
end

local function skipped(reason)
    return { ok = false, error = "skipped: " .. reason }
end

local function buildDebugPayload(bridge)
    local results = {}

    results.getCraftableItems = callBridge(bridge, "getCraftableItems", {})
    results.getConfiguration = callBridge(bridge, "getConfiguration")
    results.getName = callBridge(bridge, "getName")
    results.getCells = callBridge(bridge, "getCells")
    results.getDrives = callBridge(bridge, "getDrives")
    results.getCraftingTasks = callBridge(bridge, "getCraftingTasks")
    results.getCraftingCPUs = callBridge(bridge, "getCraftingCPUs")
    results.getPatterns = callBridge(bridge, "getPatterns")

    if DEBUG_TASK_ID then
        results.getCraftingTask = callBridge(bridge, "getCraftingTask", DEBUG_TASK_ID)
    else
        results.getCraftingTask = skipped("DEBUG_TASK_ID not set")
    end

    if DEBUG_ITEM_ID ~= "" then
        results.isCraftable = callBridge(bridge, "isCraftable", { name = DEBUG_ITEM_ID })
        results.isCrafting = callBridge(bridge, "isCrafting", { name = DEBUG_ITEM_ID })
        results.getItem = callBridge(bridge, "getItem", { name = DEBUG_ITEM_ID })
    else
        results.isCraftable = skipped("DEBUG_ITEM_ID not set")
        results.isCrafting = skipped("DEBUG_ITEM_ID not set")
        results.getItem = skipped("DEBUG_ITEM_ID not set")
    end

    return {
        timestamp = util.nowMs(),
        results = results
    }
end

local function sendLoop(ws)
    local bridge = nil
    while true do
        bridge = aeBridge.ensureBridge(bridge)

        local payload
        if bridge then
            payload = packets.bridgeDebug(config.DEVICE_ID, buildDebugPayload(bridge))
        else
            payload = packets.bridgeDebug(config.DEVICE_ID, {
                timestamp = util.nowMs(),
                results = {
                    bridge = { ok = false, error = "no me_bridge" }
                }
            })
        end

        util.sendJson(ws, payload)
        util.sleepSeconds(SEND_INTERVAL_SECONDS)
    end
end

local function runSession(ws)
    print("Bridge debug online.")
    sendLoop(ws)
end

wsClient.run(config.WS_URL, runSession, config.RECONNECT_DELAY)
