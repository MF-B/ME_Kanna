local M = {}

function M.sleepSeconds(seconds)
    sleep(seconds or 0)
end

function M.sendJson(ws, payload)
    ws.send(textutils.serializeJSON(payload))
end

function M.nowMs()
    if os.epoch then
        return os.epoch("utc")
    end
    return math.floor(os.clock() * 1000)
end

return M
