-- ================= 配置区域 =================
local API_URL = "http://192.168.1.34:8080/config/whitelist"
local WS_URL  = "ws://192.168.1.34:8080/ws/minecraft" 
local DEVICE_ID = "ae_hub"
-- ===========================================

local ae_device = nil
local MONITORED_ITEMS = {} 

local function findBridge()
    print("Scanning for ME Bridge...")
    for _, name in pairs(peripheral.getNames()) do
        if peripheral.getType(name) == "me_bridge" then
            print(" + Linked: " .. name)
            return peripheral.wrap(name)
        end
    end
    return nil
end

local function syncConfig()
    print("Syncing config...")
    local response = http.get(API_URL)
    if response then
        local body = response.readAll()
        response.close()
        local data = textutils.unserializeJSON(body)
        if data and data.monitored_items then
            MONITORED_ITEMS = {}
            for _, item in pairs(data.monitored_items) do
                MONITORED_ITEMS[item] = true
            end
            print("Config Synced! Count: " .. #data.monitored_items)
        end
    else
        print("Config Sync Failed!")
    end
end

local function sendLoop(ws)
    while true do
        -- 掉线重连设备
        if not ae_device then
            ae_device = findBridge()
            if not ae_device then sleep(5) end
        end

        if ae_device then
            local filtered_items = {}
            
            local all_items = ae_device.getItems() 
            
            for _, item in pairs(all_items) do
                local count = item.count or 0
                local name = item.name
                
                if name and MONITORED_ITEMS[name] then
                    filtered_items[name] = (filtered_items[name] or 0) + count
                end
            end

            local payload = {
                type = "update",
		id = DEVICE_ID,
                data = {
                    [DEVICE_ID] = {
                        name = "Main Storage",
                        isActive = true,
                        raw_items = filtered_items
                    }
                }
            }
            ws.send(textutils.serializeJSON(payload))
        end
        sleep(5) 
    end
end

local function receiveLoop(ws)
    while true do
        local msg = ws.receive()
        if not msg then break end
    end
end

-- ================= 主程序 =================
while true do
    term.clear()
    term.setCursorPos(1,1)
    
    syncConfig()
    ae_device = findBridge()
    
    print("Connecting WS...")
    local ws, err = http.websocket(WS_URL)
    
    if ws then
        print("Online.")
        pcall(function()
            parallel.waitForAny(
                function() sendLoop(ws) end,
                function() receiveLoop(ws) end
            )
        end)
        ws.close()
    else
        print("Offline: " .. tostring(err))
    end
    
    print("Retry in 5s...")
    sleep(5)
end
