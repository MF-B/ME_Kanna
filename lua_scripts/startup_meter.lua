-- ================= 配置 =================
local HOST = "http://127.0.0.1:8080" -- 你的 Go 后端地址
local FILES = {
    "meter.lua",
    "lib/config.lua",
    "lib/util.lua",
    "lib/ws_client.lua",
    "lib/packets.lua"
}
-- =======================================

term.clear()
term.setCursorPos(1,1)
print("=== Turtle Bootloader ===")

local function ensureDirs(path)
    local dir = fs.getDir(path)
    if dir ~= "" and not fs.exists(dir) then
        fs.makeDir(dir)
    end
end

local function fetchFile(path)
    local updateUrl = HOST .. "/lua/" .. path
    local response = http.get(updateUrl)
    if not response then
        print("Update Failed: " .. path .. " (Server offline?)")
        return false
    end

    local content = response.readAll()
    response.close()

    ensureDirs(path)
    local file = fs.open(path, "w")
    file.write(content)
    file.close()

    print("Update Success: " .. path)
    return true
end

local function updateCode()
    print("Checking for updates...")
    for _, path in ipairs(FILES) do
        if not fetchFile(path) then
            return false
        end
    end
    return true
end

-- 没文件就死循环下载，有文件就尝试更新
if fs.exists("meter.lua") then
    updateCode()
else
    while not updateCode() do
        print("Retrying in 5s...")
        sleep(5)
    end
end

print("Launching meter.lua...")
sleep(1)

-- 运行海龟主程序
local ok, err = pcall(function()
    shell.run("meter.lua")
end)

if not ok then
    print("CRASHED: " .. tostring(err))
    print("Rebooting in 10s...")
    sleep(10)
    os.reboot()
end
