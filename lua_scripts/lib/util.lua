local M = {}

function M.sleepSeconds(seconds)
    sleep(seconds or 0)
end

function M.sendJson(ws, payload)
    ws.send(textutils.serializeJSON(payload))
end

return M
