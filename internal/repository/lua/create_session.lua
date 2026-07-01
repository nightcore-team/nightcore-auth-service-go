local userSessionsKey = KEYS[1]
local sessionKey = KEYS[2]
local refreshToken = ARGV[1]
local ttl = tonumber(ARGV[2])
local maxSessions = tonumber(ARGV[3])
local sessionData = ARGV[4]

if redis.call('SCARD', userSessionsKey) >= maxSessions then
    local members = redis.call('SMEMBERS', userSessionsKey)
    for _, token in ipairs(members) do
        redis.call('DEL', 'session:' .. token)
    end
    redis.call('DEL', userSessionsKey)
end

redis.call('SADD', userSessionsKey, refreshToken)
redis.call('EXPIRE', userSessionsKey, ttl)
redis.call('SET', sessionKey, sessionData, 'EX', ttl)

return 1
