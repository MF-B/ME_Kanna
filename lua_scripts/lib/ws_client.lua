local util = require("util")

local M = {}

function M.run(wsUrl, onSession, reconnectDelay)
    while true do
        term.clear()
        term.setCursorPos(1, 1)
        print("Connecting...")

        local ws, err = http.websocket(wsUrl)
        if ws then
            local ok, runErr = pcall(onSession, ws)
            if not ok then
                print("Session error: " .. tostring(runErr))
            end
            ws.close()
        else
            print("Offline: " .. tostring(err))
        end

        print("Reconnecting in " .. tostring(reconnectDelay) .. "s...")
        util.sleepSeconds(reconnectDelay)
    end
end

return M
