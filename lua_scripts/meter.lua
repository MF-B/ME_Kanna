-- ================= 配置区域 =================
local WS_URL = "ws://192.168.1.34:8080/ws/minecraft" 
local FACTORY_ID = "iron_farm"
local SIDE_REDSTONE = "top"
-- ===========================================

local function mainLoop(ws)
    print("Turtle Online: " .. FACTORY_ID)
    
    while true do
        -- 并行处理：既要发数据，又要收指令
        parallel.waitForAny(
            -- 任务1: 监控库存 (发送)
            function()
                while true do
                    local event, side = os.pullEvent("turtle_inventory")
                    local item = turtle.getItemDetail(1)
                    if item then
                        local payload = {
                            type = "production_flow",
                            id = FACTORY_ID, -- 这个 ID 很重要，后端靠它认人
                            delta = item.count,
                            item = item.name
                        }
                        ws.send(textutils.serializeJSON(payload))
                        while not turtle.dropDown() do sleep(2) end
                    end
                end
            end,

            -- 任务2: 监听指令 (接收)
            function()
                while true do
                    -- 直接从 WebSocket 收消息
                    local msg = ws.receive()
                    if msg then
                        local cmd = textutils.unserializeJSON(msg)
                        if cmd and cmd.action then
                            print("CMD: " .. cmd.action)
                            if cmd.action == "stop" then
                                rs.setOutput(SIDE_REDSTONE, true)
                                print("-> STOPPED")
                            elseif cmd.action == "start" then
                                rs.setOutput(SIDE_REDSTONE, false)
                                print("-> STARTED")
                            end
                        end
                    else
                        break -- 连接断开
                    end
                end
            end
        )
        break 
    end
end

while true do
    term.clear()
    term.setCursorPos(1,1)
    print("Connecting...")
    local ws, err = http.websocket(WS_URL)
    if ws then
        pcall(mainLoop, ws)
        ws.close()
    else
        print("Error: " .. tostring(err))
    end
    print("Reconnecting in 5s...")
    sleep(5)
end
