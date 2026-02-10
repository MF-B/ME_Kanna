package.path = "lib/?.lua;" .. package.path

local config = require("config")
local wsClient = require("ws_client")
local packets = require("packets")
local util = require("util")

local function mainLoop(ws)
    print("Turtle Online: " .. config.FACTORY_ID)
    
    while true do
        -- 并行处理：既要发数据，又要收指令
        parallel.waitForAny(
            -- 任务1: 监控库存 (发送)
            function()
                while true do
                    local event, side = os.pullEvent("turtle_inventory")
                    local item = turtle.getItemDetail(1)
                    if item then
                        local payload = packets.productionFlow(config.FACTORY_ID, item.count, item.name)
                        util.sendJson(ws, payload)
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
                                rs.setOutput(config.SIDE_REDSTONE, true)
                                print("-> STOPPED")
                            elseif cmd.action == "start" then
                                rs.setOutput(config.SIDE_REDSTONE, false)
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

wsClient.run(config.WS_URL, mainLoop, config.RECONNECT_DELAY)
