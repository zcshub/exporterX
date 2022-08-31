local MonsterData = require("hook.LogicProcess.MonsterData")

local cacheData = {}
local changedData = {}

-- 要缓存的数据列表
function GlobalCacheDataList()
    return {
        "MonsterData",
        "MonsterTemplateData",

        "TaskData",

    }
end

local function dump(o)
    if type(o) == 'table' then
       local s = '{ '
       for k,v in pairs(o) do
          if type(k) ~= 'number' then k = '"'..k..'"' end
          s = s .. '['..k..'] = ' .. dump(v) .. ','
       end
       return s .. '} '
    else
       return tostring(o)
    end
 end

-- 接收缓存数据
function ReceiveCacheData(name, table)
    cacheData[name] = table
end

-- 处理缓存数据
function ProcessCacheData()
    local monsterData = MonsterData.replaceTemplateData(cacheData["MonsterData"], cacheData["MonsterTemplateData"])
    changedData["MonsterData"] = monsterData
end

-- 获取改变的数据
function GetChangedData()
    return changedData
end

