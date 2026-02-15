package.path = "lib/?.lua;" .. package.path

local config = require("config")
local wsClient = require("ws_client")
local whitelist = require("whitelist")
local aeBridge = require("ae_bridge")
local packets = require("packets")
local util = require("util")

local ae_device = nil

local function normalizeCraftables(rawCraftables, bridge)
    local normalizedCraftables = {}
    if type(rawCraftables) ~= "table" then return normalizedCraftables end

    for _, craftableEntry in ipairs(rawCraftables) do
        local itemId = craftableEntry and craftableEntry.name
        if type(itemId) == "string" and itemId ~= "" then
            -- Query current stock from ME system
            local count = 0
            if bridge then
                local detail = bridge.getItem({name = itemId})
                count = (detail and detail.count) or 0
            end
            table.insert(normalizedCraftables, {
                itemId = itemId,
                itemName = craftableEntry.displayName or itemId,
                count = count
            })
        end
    end

    return normalizedCraftables
end

local function sendLoop(ws)
    while true do
        ae_device = aeBridge.ensureBridge(ae_device)

        if ae_device then
            -- 1. 采集数据 (全都在 aeBridge 里封装好了)
            local filteredItems = aeBridge.collectFilteredItems(ae_device, whitelist.getList())
            local energy = aeBridge.collectEnergy(ae_device)
            local storage = aeBridge.collectStorage(ae_device)
            
            -- 2. 打包发送 (whitelist.getVersion() 放在这里)
            local payload = packets.inventoryUpdate(
                config.DEVICE_ID,
                "Main Storage",
                true,
                filteredItems,
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
            elseif packet and packet.type == "cmd_craftables" then
                ae_device = aeBridge.ensureBridge(ae_device)
                if ae_device then
                    local craftableList = normalizeCraftables(aeBridge.getCraftables(ae_device), ae_device)
                    util.sendJson(ws, packets.craftablesUpdate(config.DEVICE_ID, craftableList, packet.requestId))
                    print("Craftables sent: " .. tostring(#craftableList))
                else
                    print("Craftables request failed: no ME Bridge")
                end
            elseif packet and packet.type == "craft" then
                local itemId = packet.itemId
                local count = tonumber(packet.count) or 1
                if itemId and itemId ~= "" then
                    ae_device = aeBridge.ensureBridge(ae_device)
                    if ae_device then
                        local task, err = aeBridge.craft(ae_device, itemId, count)
                        local ok = task ~= nil
                        if ok then
                            local craftId = task.id or task.taskId or "?"
                            print("Craft queued: " .. tostring(itemId) .. " x" .. tostring(count) .. " (id=" .. tostring(craftId) .. ")")
                        else
                            print("Craft failed: " .. tostring(itemId) .. " x" .. tostring(count) .. " reason=" .. tostring(err))
                        end
                    else
                        print("Craft failed: no ME Bridge")
                    end
                else
                    print("Craft ignored: missing item id")
                end
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