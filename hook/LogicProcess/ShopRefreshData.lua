local ShopRefreshData = {}

function ShopRefreshData.handle(PreData)
    local maxTimes = 0
    -- 确定配表中的最大刷新次数
    for k, v in pairs(PreData) do
        local intK = tonumber(k)
        if intK > maxTimes then
            maxTimes = intK
        end
    end
    -- 填充表中可能没有的次数，每一项使用前一次的数值
    local lastAvailableCost = 0
    for i = 1, maxTimes do
        if PreData[tostring(i)] == nil then
            PreData[tostring(i)] = lastAvailableCost
        else
            lastAvailableCost = PreData[tostring(i)]
        end
    end
    return PreData
end

return ShopRefreshData