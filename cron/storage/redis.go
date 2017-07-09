/**
 * Provides a Redis storage adapter for storing and retrieving Schedules.
 */

package msCronStorage

import (
	// Utilities
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	// Redis.
	"github.com/mediocregopher/radix.v2/redis"

	// Internal dependencies.
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
)

/**
 * Constants.
 */

// redisScheduleHashPrefix holds the prefix that is prepended to a Schedule's ID
// to form it's hash key in the Redis database i.e. the Schedule with ID 1 will
// be stored in a Redis hash with key "schedule:1".
const redisScheduleHashPrefix = "schedule:"

// redisScheduleStartIndex holds the key of the Sorted Set data structure that
// stores an index of the start times for all Schedules.
const redisScheduleStartIndex = "schedules_start_index"

// redisScheduleStopIndex holds the key of the Sorted Set data structure that
// stores an index of the stop times for all Schedules.
const redisScheduleStopIndex = "schedules_stop_index"

// redisScheduleIDIndex holds the key of the Sorted Set data structure that
// stores an index of the IDs for all Schedules.
const redisScheduleIDIndex = "schedules"

// redisScheduleSearchScript holds the name of the file that contains the Lua
// script that searches for and returns Schedules candidate for triggering.
const redisScheduleSearchScript = "search.lua"

/**
 * Redis storage provider.
 */

// Redis implements the Storage interface, allowing to use Redis as a Storage
// engine.
type Redis struct {
	dsn    string
	client *redis.Client
}

// Create implements Storage.Create(). It stores the given Schedule object as a
// new Hash in the Redis Storage and it returns an automatically generated ID.
func (storage Redis) Create(schedule *schedule.Schedule) (*int, error) {
	// Generate an ID and store the Schedule.
	scheduleID, err := storage.generateID()
	if err != nil {
		return nil, err
	}

	// Set the CreatedAt field, if not yet set.
	now := time.Now()
	if schedule.CreatedAt == nil {
		schedule.CreatedAt = &now
	}

	err = storage.set(*scheduleID, schedule, true)
	if err != nil {
		return nil, err
	}

	// Set the new ID in the corresponding Schedule field.
	schedule.ID = *scheduleID

	return scheduleID, nil
}

// Get implements Storage.Get(). It retrieves from Storage and returns the
// Schedule for the given ID.
func (storage Redis) Get(scheduleID int) (*schedule.Schedule, error) {
	if storage.client == nil {
		return nil, fmt.Errorf("trying to get a Schedule from the database while the Redis client has not been initialized yet")
	}

	key := redisKey(scheduleID)

	hashFields, err := storage.client.Cmd("HGETALL", key).List()
	if err != nil {
		return nil, err
	}

	// Convert the Redis hash into a Schedule object.
	schedule, err := fromHashFields(&hashFields)
	if err != nil {
		return nil, err
	}

	schedule.ID = scheduleID

	return schedule, nil
}

// Update implements Storage.Update(). It stores the given Schedule object as a Hash
// in the Redis Storage, overriding the existing fields for the Hash with the
// given ID.
func (storage Redis) Update(schedule *schedule.Schedule, updateTimestamp bool) error {
	return storage.set(schedule.ID, schedule, updateTimestamp)
}

// Search implements storage.Search(). It search for and returns Schedule
// objects that are candidates for evaluating and triggering their Watches
// within the time period starting from now (the moment the function is called)
// and ending after the given interval.
func (storage Redis) Search(pollInterval time.Duration) ([]*schedule.Schedule, error) {
	script, err := loadSearchScript()
	if err != nil {
		return nil, err
	}

	// Start and end times for the search.
	start := time.Now()
	stop := start.Add(pollInterval)

	// Get candidate Schedules using the Lua script.
	// @I Store the Lua script in Redis and trigger it by its hash
	rSchedules, err := storage.client.Cmd(
		"EVAL",
		script,
		2,
		redisScheduleStartIndex,
		redisScheduleStopIndex,
		redisScheduleHashPrefix,
		start.UnixNano(),
		stop.UnixNano(),
	).Array()
	if err != nil {
		return nil, err
	}

	var hashFields [][]string

	// Unwrap fields from the Redis response.
	for _, v := range rSchedules {
		rSchedule, err := v.Array()
		if err != nil {
			return nil, err
		}

		var tmpHashFields []string

		for _, vv := range rSchedule {
			field, err := vv.Str()
			if err != nil {
				return nil, err
			}
			tmpHashFields = append(tmpHashFields, field)
		}

		hashFields = append(hashFields, tmpHashFields)
	}

	// Convert the Hash fields into Schedule objects and return them.
	return fromHashes(hashFields)
}

// NewRedisStorage implements the StorageFactory function type. It initiates a
// connection to the Redis database defined in the given configuration, and it
// returns the Storage engine object.
var NewRedisStorage = func(config map[string]interface{}) (Storage, error) {
	dsn, ok := config["dsn"]
	if !ok {
		err := fmt.Errorf("the DSN configuration option is required for the Redis storage")
		return nil, err
	}

	sDSN := dsn.(string)

	client, err := redis.Dial("tcp", sDSN)
	if err != nil {
		err := fmt.Errorf("failed to connect to Redis: %s", err.Error())
		return nil, err
	}

	storage := Redis{
		dsn:    sDSN,
		client: client,
	}

	return storage, nil
}

/**
 * For internal use.
 */

// set stores a Schedule object into a Redis Hash at the key corresponding to
// the given ID.
func (storage Redis) set(scheduleID int, schedule *schedule.Schedule, updateTimestamp bool) error {
	if storage.client == nil {
		return fmt.Errorf("the Redis client has not been initialized yet")
	}

	// Update the UpdatedAt field.
	if updateTimestamp {
		now := time.Now()
		schedule.UpdatedAt = &now
	}

	// Convert the Schedule object into the Hash fields that will be stored.
	fields := toHashFields(schedule)

	// Store the Schedule and update the index sets.
	key := redisKey(scheduleID)
	err := storage.client.Cmd(
		"HMSET",
		key,
		*fields,
	).Err
	if err != nil {
		return err
	}

	// Set the ID in the corresponding index.
	// @I Add the Schedule's ID to the index only when creating it
	err = storage.client.Cmd("ZADD", redisScheduleIDIndex, scheduleID, key).Err
	if err != nil {
		return err
	}

	// Set the start time in the corresponding index.
	start := timeToHashField(schedule.Start)
	err = storage.client.Cmd("ZADD", redisScheduleStartIndex, start, scheduleID).Err
	if err != nil {
		return err
	}

	// Set the stop time in the corresponding index.
	stop := timeToHashField(schedule.Stop)
	err = storage.client.Cmd("ZADD", redisScheduleStopIndex, stop, scheduleID).Err
	if err != nil {
		return err
	}

	return nil
}

// generateID generates an ID for a new Schedule by incrementing the last known
// Schedule ID.
func (storage Redis) generateID() (*int, error) {
	// @I Investigate risk of a Schedule overriding another due to race conditions
	//    when creating them

	if storage.client == nil {
		return nil, fmt.Errorf("the Redis client has not been initialized yet")
	}

	// Get the last ID that exists on the Schedules index set, so that we can
	// generate the next one.
	r, err := storage.client.Cmd("ZREVRANGE", redisScheduleIDIndex, 0, 0, "WITHSCORES").List()
	if err != nil {
		return nil, err
	}

	// If there are no Schedules yet, start with ID 1.
	if len(r) == 0 {
		newScheduleID := 1
		return &newScheduleID, nil
	}

	latestScheduleID, err := strconv.Atoi(r[1])
	if err != nil {
		return nil, err
	}

	newScheduleID := latestScheduleID + 1
	return &newScheduleID, nil
}

// redisKey generates a Redis key for the given Schedule's ID.
func redisKey(scheduleID int) string {
	return redisScheduleHashPrefix + strconv.Itoa(scheduleID)
}

// toHashFields converts a Schedule object into an array of key/value fields
// ready to be stored in a Redis Hash data structure.
func toHashFields(schedule *schedule.Schedule) *[]interface{} {
	var hashFields []interface{}

	// Mandatory fields.
	// WatchesIDs.
	hashFields = append(hashFields, "watches_ids")
	hashFields = append(hashFields, idsToHashField(&schedule.WatchesIDs))
	// Interval.
	hashFields = append(hashFields, "interval")
	hashFields = append(hashFields, schedule.Interval.Nanoseconds())
	// Enabled.
	hashFields = append(hashFields, "enabled")
	hashFields = append(hashFields, schedule.Enabled)

	// Optional fields.
	// Start.
	if schedule.Start != nil {
		hashFields = append(hashFields, "start")
		hashFields = append(hashFields, schedule.Start.UnixNano())
	}
	// Stop.
	if schedule.Stop != nil {
		hashFields = append(hashFields, "stop")
		hashFields = append(hashFields, schedule.Stop.UnixNano())
	}
	// Last.
	if schedule.Last != nil {
		hashFields = append(hashFields, "last")
		hashFields = append(hashFields, schedule.Last.UnixNano())
	}
	// CreatedAt.
	if schedule.CreatedAt != nil {
		hashFields = append(hashFields, "created_at")
		hashFields = append(hashFields, schedule.CreatedAt.UnixNano())
	}
	// UpdatedAt.
	if schedule.UpdatedAt != nil {
		hashFields = append(hashFields, "updated_at")
		hashFields = append(hashFields, schedule.UpdatedAt.UnixNano())
	}

	return &hashFields
}

// fromHashFields converts an array holding the key/value fields of a Redis Hash
// data structure into a Schedule object.
func fromHashFields(hash *[]string) (*schedule.Schedule, error) {
	kvHash := make(map[string]string)
	key := ""
	for _, v := range *hash {
		if key == "" {
			key = v
		} else {
			kvHash[key] = v
			key = ""
		}
	}

	schedule := schedule.Schedule{}
	var err error

	// ID. It is not available as a Hash field when we get the Hash individual,
	// but we make it available in the Lua script where we return multiple Hashes.
	if v, ok := kvHash["id"]; ok {
		schedule.ID, err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	}

	// Mandatory fields.
	// WatchesIDs.
	schedule.WatchesIDs, err = idsFromHashField(kvHash["watches_ids"])
	if err != nil {
		return nil, err
	}
	// Interval.
	interval, err := durationFromHashField(kvHash["interval"])
	if err != nil {
		return nil, err
	}
	schedule.Interval = *interval
	// Enabled.
	enabled, err := boolFromHashField(kvHash["enabled"])
	// @I Add the Schedule's ID to the error message when the enabled field holds
	//    a wrong value in Redis
	if err != nil {
		return nil, err
	}
	schedule.Enabled = *enabled

	// Optional fields.
	// Start.
	if v, ok := kvHash["start"]; ok {
		schedule.Start, err = timeFromHashField(v)
		if err != nil {
			return nil, err
		}
	}
	// Stop.
	if v, ok := kvHash["stop"]; ok {
		schedule.Stop, err = timeFromHashField(v)
		if err != nil {
			return nil, err
		}
	}
	// Last.
	if v, ok := kvHash["last"]; ok {
		schedule.Last, err = timeFromHashField(v)
		if err != nil {
			return nil, err
		}
	}
	// CreatedAt.
	if v, ok := kvHash["created_at"]; ok {
		schedule.CreatedAt, err = timeFromHashField(v)
		if err != nil {
			return nil, err
		}
	}
	// UpdatedAt.
	if v, ok := kvHash["updated_at"]; ok {
		schedule.UpdatedAt, err = timeFromHashField(v)
		if err != nil {
			return nil, err
		}
	}

	return &schedule, nil
}

// fromHashes converts an array of Redis Hashes (given as an array of strings
// i.e. array of array of strings) into an array of Schedule objects.
func fromHashes(hashes [][]string) ([]*schedule.Schedule, error) {
	schedules := make([]*schedule.Schedule, len(hashes))

	for i, hash := range hashes {
		schedule, err := fromHashFields(&hash)
		if err != nil {
			return nil, err
		}
		schedules[i] = schedule
	}

	return schedules, nil
}

// idsToHashField converts an array of integer IDs as stored in a Schedule
// object field into a string containing them as concatenated string values, as
// required for storing them as a field in a Redis Hash data structure.
func idsToHashField(aIntIDs *[]int) string {
	aStringIDs := make([]string, len(*aIntIDs))
	for k, v := range *aIntIDs {
		aStringIDs[k] = strconv.Itoa(v)
	}
	return strings.Join(aStringIDs, ",")
}

// idsFromHashField converts a string that contains concatenated integer IDs
// as stored in a Redis Hash field into an array of integer IDs, as required for
// storing them as a field in a Schedule object.
func idsFromHashField(sStringIDs string) ([]int, error) {
	aStringIDs := strings.Split(sStringIDs, ",")
	aIntIDs := make([]int, len(aStringIDs))
	var err error
	for i, ID := range aStringIDs {
		aIntIDs[i], err = strconv.Atoi(ID)
		if err != nil {
			return nil, err
		}
	}

	return aIntIDs, nil
}

// durationFromHashField converts a string that contains a duration as stored in
// a Redis Hash field into an time.Duration object, as required for storing it
// as a field in a Schedule object.
func durationFromHashField(sDuration string) (*time.Duration, error) {
	iDuration, err := strconv.ParseInt(sDuration, 10, 64)
	if err != nil {
		return nil, err
	}
	tDuration := time.Duration(iDuration)
	return &tDuration, nil
}

// boolFromHashField converts a string that contains a boolean value as stored
// in a Redis Hash field into a boolean primitive, as required for storing it as
// a field in a Schedule object.
func boolFromHashField(sBool string) (*bool, error) {
	var bBool bool
	if sBool == "1" {
		bBool = true
	} else if sBool == "0" {
		bBool = false
	} else {
		return nil, fmt.Errorf("non boolean value stored in the \"enabled\" field for the Schedule with ID")
	}

	return &bBool, nil
}

// timeToHashField converts a time.Time object into an int64 value, as required
// for storing it as a field in a Redis Hash structure.
func timeToHashField(time *time.Time) int64 {
	var iTime int64
	if time == nil {
		iTime = 0
	} else {
		iTime = time.UnixNano()
	}
	return iTime
}

// timeFromHashField converts a time string as contained in a Redis Hash field
// to a time.Time object, as required for storing it as a field in a Schedule
// object.
func timeFromHashField(sTime string) (*time.Time, error) {
	start, err := strconv.ParseInt(sTime, 10, 64)
	if err != nil {
		return nil, err
	}
	seconds := start / 1000000000
	nanoseconds := start % 1000000000
	tTime := time.Unix(seconds, nanoseconds)
	return &tTime, nil
}

// loadSearchScript loads the Lua script that searches for candidate Schedules
// from the containing file into a string that will be passed to the Redis EVAL
// command.
func loadSearchScript() ([]byte, error) {
	_, thisFilename, _, _ := runtime.Caller(1)
	filename := path.Join(path.Dir(thisFilename), redisScheduleSearchScript)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
