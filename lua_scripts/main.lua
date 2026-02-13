package.path = "lib/?.lua;" .. package.path

local config = require("config")
local wsClient = require("ws_client")
local whitelist = require("whitelist")
local aeBridge = require("ae_bridge")
local packets = require("packets")
local util = require("util")

local ae_device = nil

local function sendLoop(ws)
    while true do
        ae_device = aeBridge.ensureBridge(ae_device)

        if ae_device then
            -- 1. 采集数据 (全都在 aeBridge 里封装好了)
            local filtered_items = aeBridge.collectFilteredItems(ae_device, whitelist.getList())
            local energy = aeBridge.collectEnergy(ae_device)
            local storage = aeBridge.collectStorage(ae_device)
            
            -- 2. 打包发送 (whitelist.getVersion() 放在这里)
            local payload = packets.inventoryUpdate(
                config.DEVICE_ID,
                "Main Storage",
                true,
                filtered_items,
                energy,
                storage,
                whitelist.getVersion()
            )
            util.sendJson(ws, payload)
            
            -- 3. 检查同步 (放在发送之后，避免 HTTP 卡顿影响数据上报)
            whitelist.sync(config.API_URL, config.SYNC_INTERVAL)
        end

        util.sleepSeconds(0.5)
    end
end

local function receiveLoop(ws)
    while true do
        local msg = ws.receive() 
        if msg then
            local packet = textutils.unserializeJSON(msg)
            if whitelist.handlePacket(packet) then
                print("WS: Config Updated!")
            end
        else
            print("WS Disconnected")
            break
        end
    end
end

local function runSession(ws)
    -- 启动时强制同步一次
    whitelist.sync(config.API_URL, config.SYNC_INTERVAL, true)
    
    print("System Online.")
    parallel.waitForAny(
        function() sendLoop(ws) end,
        function() receiveLoop(ws) end
    )
end

wsClient.run(config.WS_URL, runSession, config.RECONNECT_DELAY)