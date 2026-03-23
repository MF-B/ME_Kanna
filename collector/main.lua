package.path = "lib/?.lua;" .. package.path

local config     = require("config")
local wsClient   = require("ws_client")
local aeBridge   = require("ae_bridge")
local packets    = require("packets")
local util       = require("util")
local cmdHandler = require("cmd_handler")

local ae_device = nil

local function sendLoop(ws)
    while true do
        ae_device = aeBridge.ensureBridge(ae_device)

        if ae_device then
            local ok, err = pcall(function()
                local items = aeBridge.collectItems(ae_device)
                local cpus = aeBridge.collectCPUs(ae_device)
                local energy = aeBridge.collectEnergy(ae_device)
                local storage = aeBridge.collectStorage(ae_device)

                local payload = packets.evtTick(items, cpus, energy, storage)
                util.sendJson(ws, payload)
            end)

            if not ok then
                print("Send error: " .. tostring(err))
                ae_device = nil -- bridge 可能已断开，下次重新扫描
            end
        end

        sleep(1)
    end
end

-- 构建命令分发上下文
local function makeContext(ws)
    return {
        ws = ws,
        getBridge = function() return ae_device end,
        setBridge = function(b) ae_device = b end,
    }
end

local function receiveLoop(ws)
    local ctx = makeContext(ws)

    while true do
        local msg = ws.receive(30) -- 30秒超时，防止半开连接永久阻塞
        if msg then
            local ok, packet = pcall(textutils.unserializeJSON, msg)
            if ok and type(packet) == "table" then
                cmdHandler.dispatch(packet, ctx)
            end
        else
            -- receive 返回 nil: 超时或断开
            -- 发送心跳探测连接是否存活
            local pingOk = pcall(function()
                util.sendJson(ws, {type = "ping"})
            end)
            if not pingOk then
                print("WS Disconnected")
                break
            end
        end
    end
end

local function craftEventLoop(ws)
    while true do
        local _, isError, taskId, message = os.pullEvent("ae_crafting")
        print("[CraftEvent] id=" .. tostring(taskId) .. " err=" .. tostring(isError) .. " msg=" .. tostring(message))
        pcall(function()
            util.sendJson(ws, packets.evtCrafting(taskId, isError, message))
        end)
    end
end

local function runSession(ws)
    print("System Online.")

    parallel.waitForAny(
        function() sendLoop(ws) end,
        function() receiveLoop(ws) end,
        function() craftEventLoop(ws) end
    )
end

wsClient.run(config.WS_URL, runSession, config.RECONNECT_DELAY)