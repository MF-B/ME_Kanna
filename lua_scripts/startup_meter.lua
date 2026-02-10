-- ================= 配置 =================
local HOST = "http://192.168.1.34:8080" -- 你的 Go 后端地址
local SCRIPT_NAME = "meter.lua"         -- 【重点】这里要改成海龟的脚本名！
local UPDATE_URL = HOST .. "/lua/" .. SCRIPT_NAME
-- =======================================

term.clear()
term.setCursorPos(1,1)
print("=== Turtle Bootloader ===")

local function updateCode()
    print("Checking for updates...")
    
    local response = http.get(UPDATE_URL)
    
    if response then
        local content = response.readAll()
        response.close()
        
        local file = fs.open(SCRIPT_NAME, "w")
        file.write(content)
        file.close()
        
        print("Update Success: " .. SCRIPT_NAME)
        return true
    else
        print("Update Failed! (Server offline?)")
        return false
    end
end

-- 没文件就死循环下载，有文件就尝试更新
if fs.exists(SCRIPT_NAME) then
    updateCode()
else
    while not updateCode() do
        print("Retrying in 5s...")
        sleep(5)
    end
end

print("Launching " .. SCRIPT_NAME .. "...")
sleep(1)

-- 运行海龟主程序
local ok, err = pcall(function()
    shell.run(SCRIPT_NAME)
end)

if not ok then
    print("CRASHED: " .. tostring(err))
    print("Rebooting in 10s...")
    sleep(10)
    os.reboot()
end
