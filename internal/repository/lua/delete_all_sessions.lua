local userSessionsKey = KEYS[1]
local members = redis.call('SMEMBERS', userSessionsKey)
for _, token in ipairs(members) do
    redis.call('DEL', 'session:' .. token)
end
redis.call('DEL', userSessionsKey)
return #members
