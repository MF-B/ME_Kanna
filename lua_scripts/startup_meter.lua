-- ================= Turtle Bootloader =================
local HOST = "http://127.0.0.1:8080"
local TARGET = "meter.lua"
local FILES = {
    "meter.lua",
    "lib/config.lua",
    "lib/util.lua",
    "lib/ws_client.lua",
    "lib/packets.lua",
    "lib/bootloader.lua"
}
-- ==================================================

-- 初始引导：确保 bootloader.lua 存在
local function ensureDirs(path)
    local dir = fs.getDir(path)
    if dir ~= "" and not fs.exists(dir) then
        fs.makeDir(dir)
    end
end

local function fetchOne(path)
    local resp = http.get(HOST .. "/lua/" .. path)
    if not resp then return false end
    local content = resp.readAll()
    resp.close()
    ensureDirs(path)
    local f = fs.open(path, "w")
    f.write(content)
    f.close()
    return true
end

-- 如果 bootloader 还不存在，先下载它
if not fs.exists("lib/bootloader.lua") then
    ensureDirs("lib/bootloader.lua")
    while not fetchOne("lib/bootloader.lua") do
        print("Waiting for server...")
        sleep(5)
    end
end

package.path = "/lib/?.lua;" .. package.path
local bootloader = require("bootloader")
bootloader.boot(TARGET, FILES, "Turtle Bootloader")
