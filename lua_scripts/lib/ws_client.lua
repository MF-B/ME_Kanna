local M = {}

function M.run(wsUrl, onSession, reconnectDelay)
    local firstRun = true
    while true do
        if firstRun then
            term.clear()
            term.setCursorPos(1, 1)
            firstRun = false
        end
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
        sleep(reconnectDelay)
    end
end

return M
