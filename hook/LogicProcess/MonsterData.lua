local M = {}


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

local getFuncData = function(data, lv)
    if type(data) == 'table' then
        if data[1] == 2 then
            -- Switch
            for _, rangeList in ipairs(data[2]) do
                if lv >= rangeList[1] and lv <= rangeList[2] then
                    if type(rangeList[3]) == "table" then
                        local val = 0
                        for _, num in ipairs(rangeList[3]) do
                            val = val * lv + num
                        end
                        return val
                    else
                        return rangeList[3]
                    end
                end
            end
            return data
        elseif data[1] == 4 then
            -- Func1    
            local val = 0
            for _, num in ipairs(data[2]) do
                val = val * lv + num
            end
            return val
        elseif data[1] == 3 then
            return data
        else
            error(dump(data))
        end
    else
        return data
    end
    return data
end

function M.replaceTemplateData(monsterData, mongterTemplateData)
    for key, value in pairs(monsterData) do
        if value["Template"] ~= "" then
            local templateData = mongterTemplateData[value["Template"]]
            for k, v in pairs(templateData) do
                if k ~= 'Id' then
                    value[k] = getFuncData(v, value["Level"])
                end
            end
        end
    end
    return monsterData
end

return M