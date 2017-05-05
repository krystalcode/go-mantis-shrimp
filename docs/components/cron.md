# Cron Component

The Cron component provides a way to schedule triggering of Watches at regular intervals, just like cron on Unix systems. At regular intervals, configured to 1 second by default, the Storage Adapter searches for Watches that are eligible for triggering between that time and the time of the next search, and it passes found Watches to the program. The program then filters the Watches by any additional conditions necessary and triggers them by making calls to the Watch API.

Please note that this component is currently used for development purposes and it has not been tested for production systems yet. It is most likely that a separate component will be developed for production use, one that supports distributed scheduling via a library such as [Dkron](http://dkron.io/). This component might be further developed as well if there is community interest.

## Redis Implementation

The currently supported Storage Adapter stores Schedules on a Redis datastore. Schedules are units that define the scheduling of one or more Watches at regular intervals, with start and end points. The start and end points of all Schedules are indexed in Redis Sorted Sets everytime Schedules are created or updated. When the Storage Adapter searches for candidate Watches, it gets all Watches with a start point less or equal to the present time, all Watches with an end point larger than the next time the search will be run, and it creates the union of these two sets of Watches. It then removes from the union set all Watches that are disabled, and Watches that have their next execution time falling outside of the search interval. This logic, executed by a Lua script inside Redis for performance reasons, results in the set of all Watches that should be executed between the present and the next search point. The Watches are then returned to the main cron programe for triggering.

The search interval, therefore, define the resolution with which Watches are triggered. The default setting is 1 second.

The Redis datastore should be considered to persist its data, if persistence is required.
