package.path = "lib/?.lua;" .. package.path

local config = require("config")
local wsClient = require("ws_client")
local packets = require("packets")
local util = require("util")

local function ensureLabel()
    local label = os.getComputerLabel()
    if label and label ~= "" then
        return label
    end

    write("Enter factory name (label): ")
    local input = read()
    if input and input ~= "" then
        os.setComputerLabel(input)
        return input
    end

    return "turtle_" .. tostring(os.getComputerID())
end

local function resolveFactoryId(label)
    if config.FACTORY_ID and config.FACTORY_ID ~= "" then
        return config.FACTORY_ID
    end
    if label and label ~= "" then
        return label
    end
    return "turtle_" .. tostring(os.getComputerID())
end

local function resolveFactoryName(factoryId, label)
    if label and label ~= "" then
        return label
    end
    return factoryId
end

local function mainLoop(ws)
    local label = ensureLabel()
    local factoryId = resolveFactoryId(label)
    local factoryName = resolveFactoryName(factoryId, label)
    print("Turtle Online: " .. factoryId)
    
    while true do
        -- 并行处理：既要发数据，又要收指令
        parallel.waitForAny(
            -- 任务1: 监控库存 (发送)
            function()
                while true do
                    local _, _ = os.pullEvent("turtle_inventory")
                    local itemDetail = turtle.getItemDetail(1)
                    if itemDetail then
                        local itemId = itemDetail.name
                        local payload = packets.productionFlow(factoryId, factoryName, itemDetail.count, itemId)
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
