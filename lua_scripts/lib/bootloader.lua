-- ================= 通用 Bootloader =================
-- 用法: 在 startup.lua 中设置 TARGET 和 FILES 后 dofile 此文件
-- 或直接在两个 startup 文件中调用 boot(target, files)

local HOST = "http://127.0.0.1:8080"

local function ensureDirs(path)
    local dir = fs.getDir(path)
    if dir ~= "" and not fs.exists(dir) then
        fs.makeDir(dir)
    end
end

local function fetchFile(host, path)
    local updateUrl = host .. "/lua/" .. path
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

local function updateCode(host, files)
    print("Checking for updates...")
    for _, path in ipairs(files) do
        if not fetchFile(host, path) then
            return false
        end
    end
    return true
end

local function boot(target, files, title)
    term.clear()
    term.setCursorPos(1, 1)
    print("=== " .. (title or "MineCCT Bootloader") .. " ===")

    if fs.exists(target) then
        updateCode(HOST, files)
    else
        while not updateCode(HOST, files) do
            print("Retrying in 5s...")
            sleep(5)
        end
    end

    print("Launching " .. target .. "...")
    sleep(1)

    local ok, err = pcall(function()
        shell.run(target)
    end)

    if not ok then
        print("CRASHED: " .. tostring(err))
        print("Rebooting in 10s...")
        sleep(10)
        os.reboot()
    end
end

return { boot = boot }
