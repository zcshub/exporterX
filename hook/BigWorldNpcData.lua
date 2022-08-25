
local stringSplit = function(input, sep)
    local pos,arr = 1, {}
    -- for each divider found
    for st,sp in function() return string.find(input, sep, pos, true) end do
        table.insert(arr, string.sub(input, pos, st - 1))
        pos = sp + 1
    end
    arr[#arr+1] = string.sub(input, pos)
    return arr
end

local gridsToPosition = function(x, y)
	if x < 0 then
        x = -x + 10000
    end
    if y < 0 then
        y = -y + 10000
    end
    return x * 10000 * 10 + y
end


-- hook BigWorldNpcData的字段VolumePoints导出方式
function BigWorldNpcData_VolumePoints(text)
    local r = {elements = {}}
    if text == "" then
        return r
    end
    local arr = stringSplit(text, ";")
    for _, posPairText in ipairs(arr) do
        posPairText = string.sub(posPairText, 2, string.len(posPairText)-1)
        local x, y = unpack(stringSplit(posPairText, "#"))
        r.elements[#r.elements+1] = gridsToPosition(tonumber(x), tonumber(y))
    end
    return r
end


-- hook BigWorldNpcData的字段ReactPoints导出方式
function BigWorldNpcData_ReactPoints(text)
    local r = {elements = {}}
    if text == "" then
        return r
    end
    local arr = stringSplit(text, ";")
    for _, posPairText in ipairs(arr) do
        posPairText = string.sub(posPairText, 2, string.len(posPairText)-1)
        local x, y = unpack(stringSplit(posPairText, "#"))
        r.elements[#r.elements+1] = gridsToPosition(tonumber(x), tonumber(y))
    end
    return r
end

function BigWorldNpcData_Test(text)
    if text == "" then
        return 1.2
    end
    return 1
end