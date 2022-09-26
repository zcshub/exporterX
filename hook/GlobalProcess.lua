local BigWorldFogData = require("hook.LogicProcess.BigWorldFogData")
local DropAndAwardData = require("hook.LogicProcess.DropAndAwardData")
local AchievementData = require("hook.LogicProcess.AchievementData")
local ShopRefreshData = require("hook.LogicProcess.ShopRefreshData")
local MonsterData = require("hook.LogicProcess.MonsterData")

local cacheData = {}
local changedData = {}

-- 要缓存的数据列表
function GlobalCacheDataList()
    return {
        "AchievementData",
        "AchievementTabData",

        "MonsterData",
        "MonsterTemplateData",

        "TaskData",

        "ShopRefreshData",

        "DropCommonData",
        "AwardData",

        "BigWorldFogData",
        "GatherResData",
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
    if cacheData["MonsterData"] ~= nil and cacheData["MonsterTemplateData"] ~= nil then
        local monsterData = MonsterData.replaceTemplateData(cacheData["MonsterData"], cacheData["MonsterTemplateData"])
        changedData["MonsterData"] = monsterData
    end
    if cacheData["ShopRefreshData"] ~= nil then
        changedData['ShopRefreshData'] = ShopRefreshData.handle(cacheData["ShopRefreshData"])
    end
    if cacheData['AchievementData'] ~= nil and cacheData["AchievementTabData"] ~= nil then
        changedData['AchievementData'], changedData["AchievementTabData"] = AchievementData.handle(cacheData['AchievementData'], cacheData["AchievementTabData"])
    end
    if cacheData["AwardData"] ~= nil then
        changedData['AwardData'] = DropAndAwardData.handle(cacheData["AwardData"])
    end
    if cacheData["DropCommonData"] ~= nil then
        changedData['DropCommonData'] = DropAndAwardData.handle(cacheData["DropCommonData"])
    end
    if cacheData['BigWorldFogData'] ~= nil and cacheData['GatherResData'] ~= nil then
        changedData['BigWorldFogData'], changedData['GatherResData'] = BigWorldFogData.handle(cacheData['BigWorldFogData'], cacheData['GatherResData'])
    end
end

-- 获取改变的数据
function GetChangedData()
    return changedData
end

