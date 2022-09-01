local AchievementData = {}

--- [API] string split by delimiter
local function stringSplit(input, delimiter)
    input = tostring(input)
    delimiter = tostring(delimiter)
    if (delimiter=='') then return false end
    local pos,arr = 0, {}
    -- for each divider found
    for st,sp in function() return string.find(input, delimiter, pos, true) end do
        table.insert(arr, string.sub(input, pos, st - 1))
        pos = sp + 1
    end
    table.insert(arr, string.sub(input, pos))
    return arr
end


function AchievementData.handle(achievementData, achievementTabData)
    local tabTypeToTypeKey = {} -- TypeKey是achievementData的key，也就是string类型的TypeID
    for k, data in pairs(achievementTabData) do
        data['Achievements'] = {}
        tabTypeToTypeKey[data['Tab']] = k
    end

    for k, data in pairs(achievementData) do
        table.insert(achievementTabData[tabTypeToTypeKey[data['Type']]]['Achievements'], tonumber(k))
    end
    for k, data in pairs(achievementTabData) do
        table.sort(data['Achievements'])
    end
    return achievementData, achievementTabData
end

return AchievementData