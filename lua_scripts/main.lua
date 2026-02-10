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
            if os.clock() - whitelist.lastSync > config.SYNC_INTERVAL then
                whitelist.sync(config.API_URL)
            end

            local filtered_items = aeBridge.collectFilteredItems(ae_device, whitelist.isMonitored)
            local payload = packets.inventoryUpdate(config.DEVICE_ID, "Main Storage", true, filtered_items)
            util.sendJson(ws, payload)
        end

        util.sleepSeconds(5)
    end
end

local function receiveLoop(ws)
    while true do
        local msg = ws.receive()
        if not msg then break end
    end
end

local function runSession(ws)
    whitelist.sync(config.API_URL)
    ae_device = aeBridge.findBridge()

    print("Online.")
    parallel.waitForAny(
        function() sendLoop(ws) end,
        function() receiveLoop(ws) end
    )
end

-- ================= 主程序 =================
wsClient.run(config.WS_URL, runSession, config.RECONNECT_DELAY)
