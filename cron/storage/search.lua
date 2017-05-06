-- Get the Schedules that have a start time before the polling interval's start
-- time.
local start_index = redis.call("ZRANGEBYSCORE", KEYS[1], "-inf", "("..ARGV[3])
-- Get the Schedules that have a stop time after the polling interval's end
-- time.
local stop_index = redis.call("ZRANGEBYSCORE", KEYS[2], ARGV[2], "+inf")

-- @I Take into account Schedules without start and stop times.

-- Temporary table that will be holding all candidate Schedules, which is the
-- union of the tables defined above
local schedules = {}
-- Table where the final, filtered schedules that meet all conditions will be
-- held.
local filteredSchedules = {}

-- Temporary variables that hold the start and end times of the polling
-- interval.
local start = tonumber(ARGV[2])
local stop  = tonumber(ARGV[3])

-- Get the union of the Schedule candidates that meet the start and end times
-- criteria. At the same time we're creating the union of the tables, we're
-- loading the Schedules' hashes since they are values that we will be
-- returning.
-- @I Look if there's a more efficient way to get the union of two tables in Lua
for k, v in pairs(start_index) do
   local i = tonumber(v)
   schedules[i] = redis.call("HGETALL", ARGV[1]..v)
   -- Add the ID field to the returned values so that we know which Schedule the
   -- rest of the fields correspond to.
   schedules[i][#schedules[i]+1] = "id"
   schedules[i][#schedules[i]+1] = v
end
for k, v in pairs(stop_index) do
   local i = tonumber(v)
   if schedules[i] == nil then
      schedules[i] = redis.call("HGETALL", ARGV[1]..v)
      -- Add the ID field to the returned values.
      schedules[i][#schedules[i]+1] = "id"
      schedules[i][#schedules[i]+1] = v
   end
end


-- Filter the candidate Schedules:
-- - Remove disabled Schedules.
-- - Remove Schedules that their next trigger time (as indicated by their last
--   trigger time and their trigger interval) falls outside of the current
--   polling interval.
for k, v in pairs(schedules) do
   -- We copy the Schedule to a new array regardless of whether it meets the
   -- criteria, and we remove it later if it doesn't. This is not optimum, but
   -- we do it because ideally we do not need a new array to hold the filtered
   -- Schedules. Instead, we should be simply removing the Schedule from the
   -- original array, but for some reason associative arrays always return an
   -- empty value. We therefore keep the structure of the program and use an
   -- intermediary array until we find out what is wrong and fix the problem.
   -- @I Find out why associative arrays in Lua always return as empty
   filteredSchedules[#filteredSchedules+1] = schedules[k]
   local i = #filteredSchedules

   -- Create an associative array for the schedule so that we can easily get the
   -- value by key.
   local schedule = {}
   local key = ""
   for kk, vv in pairs(v) do
      if key == "" then
         key = vv
      else
         schedule[key] = vv
         key = ""
      end
   end

   -- Filter the Schedules based on the criteria described above.
   if (schedule["enabled"] == "0" or (schedule["last"] ~= nil and tonumber(schedule["last"])+tonumber(schedule["interval"]) >= stop)) then
      filteredSchedules[i] = nil
   end
end

-- @I Investigate returning values from Lua script as a MessagePack for better
--    performance
return filteredSchedules
