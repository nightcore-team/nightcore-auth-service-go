local sessionKey = KEYS[1]
local refreshToken = ARGV[1]

local data = redis.call('GET', sessionKey)
if not data then
    return false
end

redis.call('DEL', sessionKey)

local userID = data:match('"user_id":(%d+)')
if userID then
    redis.call('SREM', 'user_sessions:' .. userID, refreshToken)
end

return data
