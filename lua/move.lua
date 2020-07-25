local id = KEYS[1]
local name = KEYS[2]
local currency = KEYS[3]
local from = KEYS[4]
local fromCurrency = KEYS[5]
local to = KEYS[6]
local toCurrency = KEYS[7]
local detail = KEYS[8]

local fromAmount = tonumber(ARGV[1])
local toAmount = tonumber(ARGV[2])
local feeAmount = tonumber(ARGV[3])
local first = tonumber(ARGV[4])
local fromDB = tonumber(ARGV[5])
local toDB = tonumber(ARGV[6])

redis.call("select", fromDB)
--> 第一次调用,不检查是否有重复ID,所以这里如果输入ID重复会导致对账错误,但是概率几乎为0
if first ~= 1 then
    local setExist = redis.call("hexists", "logs", id)
    if setExist == 1 then
        return "existsed"
    end
end
--> 拼接一条记录
local msg = {}
table.insert(msg, name)
table.insert(msg, from .. "." .. fromCurrency)
table.insert(msg, to .. "." .. toCurrency)
table.insert(msg, string.format("%0.10f", fromAmount))
table.insert(msg, string.format("%0.10f", toAmount))
table.insert(msg, string.format("%0.10f", feeAmount))
table.insert(msg, string.format("%d", fromDB))
table.insert(msg, string.format("%d", toDB))
table.insert(msg, detail)
--> 减少开始用户的余额
local bFrom = tonumber(redis.call("hincrbyfloat", from, fromCurrency, 0-fromAmount))
--> from用户余额不足，回退扣除的余额，写入一个失败的msg，并返回错误"notEnoughBalance",r如果转出方是system，那么允许负的余额
if bFrom < -0.00000001 and from~="system" then
    redis.call("hincrbyfloat", from, fromCurrency, fromAmount)
    table.insert(msg, "false")
    table.insert(msg, "^")
    table.insert(msg, string.format("%0.10f", bFrom))
    redis.call("hset", "h.logs", id, table.concat(msg,"^"))
    return redis.error_reply("notEnoughBalance")
end
--> 如果存在feeAmount，那么增加对应的账户盈利
if feeAmount > 0.0 then
    redis.call("hincrbyfloat", "fee" .. name, currency, feeAmount)
end
--> 如果存在余额库转移，那么
if toDB ~= fromDB then
    redis.call("select", toDB)
end
--> 增加to的用户余额
local bTo = redis.call("hincrbyfloat", to, toCurrency, toAmount)
--> 接着写入msg
table.insert(msg, "true")
table.insert(msg, string.format("%0.10f", bFrom))
table.insert(msg, string.format("%0.10f", bTo))
--> 如果刚才切库了，得转移回来
if toDB ~= fromDB then
    --> 写入新db的msg列表
    redis.call("hset", "h.logs", id, table.concat(msg,"^"))
    redis.call("select", fromDB)
end
--> 写入当前库的msg
redis.call("hset", "h.logs", id, table.concat(msg,"^"))
-->
return msg
