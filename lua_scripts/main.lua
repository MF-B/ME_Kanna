package.path = "lib/?.lua;" .. package.path

local config = require("config")
local wsClient = require("ws_client")
local whitelist = require("whitelist")
local aeBridge = require("ae_bridge")
local packets = require("packets")
local util = require("util")

local ae_device = nil

local function collectEnergy(bridge)
    return {
        energyStored = bridge.getStoredEnergy() or 0,
        energyMax = bridge.getEnergyCapacity() or 0,
        energyUsage = bridge.getEnergyUsage() or 0,
        averageEnergyInput = bridge.getAverageEnergyInput() or 0
    }
end

local function collectStorage(bridge)
    return {
        itemTotal = bridge.getTotalItemStorage() or 0,
        itemUsed = bridge.getUsedItemStorage() or 0,
        itemAvailable = bridge.getAvailableItemStorage() or 0,
        itemExternalTotal = bridge.getTotalExternalItemStorage() or 0,
        itemExternalUsed = bridge.getUsedExternalItemStorage() or 0,
        itemExternalAvailable = bridge.getAvailableExternalItemStorage() or 0,
        fluidTotal = bridge.getTotalFluidStorage() or 0,
        fluidUsed = bridge.getUsedFluidStorage() or 0,
        fluidAvailable = bridge.getAvailableFluidStorage() or 0,
        fluidExternalTotal = bridge.getTotalExternalFluidStorage() or 0,
        fluidExternalUsed = bridge.getUsedExternalFluidStorage() or 0,
        fluidExternalAvailable = bridge.getAvailableExternalFluidStorage() or 0,
        chemicalTotal = bridge.getTotalChemicalStorage() or 0,
        chemicalUsed = bridge.getUsedChemicalStorage() or 0,
        chemicalAvailable = bridge.getAvailableChemicalStorage() or 0,
        chemicalExternalTotal = bridge.getTotalExternalChemicalStorage() or 0,
        chemicalExternalUsed = bridge.getUsedExternalChemicalStorage() or 0,
        chemicalExternalAvailable = bridge.getAvailableExternalChemicalStorage() or 0
    }
end

local function buildInventoryPayload(filtered_items, energy, storage)
    return packets.inventoryUpdate(
        config.DEVICE_ID,
        "Main Storage",
        true,
        filtered_items,
        energy,
        storage
    )
end

local function sendLoop(ws)
    while true do
        ae_device = aeBridge.ensureBridge(ae_device)

        if ae_device then
            if os.clock() - whitelist.lastSync > config.SYNC_INTERVAL then
                whitelist.sync(config.API_URL)
            end

            local filtered_items = aeBridge.collectFilteredItems(ae_device, whitelist.isMonitored)
            local energy = collectEnergy(ae_device)
            local storage = collectStorage(ae_device)
            local payload = buildInventoryPayload(filtered_items, energy, storage)
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
