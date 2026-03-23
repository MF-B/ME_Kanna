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
        local _, _ = os.pullEvent("turtle_inventory")
        local itemDetail = turtle.getItemDetail(1)
        if itemDetail then
            local payload = packets.evtProduction(factoryId, factoryName, itemDetail.name, itemDetail.count)
            util.sendJson(ws, payload)
            while not turtle.dropDown() do sleep(2) end
        end
    end
end

wsClient.run(config.WS_URL, mainLoop, config.RECONNECT_DELAY)
