# Cron Schedule

A Cron Schedule defines the execution of one or more Watches at regular intervals. The minimum interval at which the system checks for schedules that need triggering is 1s by default, but it is configurable. A Cron Schedule can optionally define a start and an end time as well. The Schedule will be triggered for the first time at the trigger interval closer to the start time, and will be regularly triggered at the defined interval until the end time.

An example of a schedule in JSON:

```
{
  // Other Schedule providers may be implemented in the future, such as
  // distributed cron.
  "type"        : "cron",
  // The interval with which the Schedule will be regularly triggered.
  "interval"    : "1m",
  // The start time of triggering the Schedule.
  "start"       : "2017-04-12T00:00:00Z",
  // The stop time of triggering the Schedule.
  "stop"        : "2017-05-12T00:00:00Z",
  // The IDs of the Watches that will be triggered.
  "watches_ids" : [12],
  // Only enabled Schedules will be triggered.
  "enabled"     : true
}
```
