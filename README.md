# go-cache

Is a simple repository to practice and demonstrate the use of channels and the async package to create an efficient thread safe cache system.

When multiple threads want to get values from the cache, but the values is not present or expired it will use a channels system so that only the first request will re-fetch the value from the databse while the others wait for a signal back. When the value is returned from the database the channels get unblocked and the values is returned properly.

Example running multiple curl commands at the same time:

```bash
read from db
read from channel
read from channel
read from db
read from channel
read from channel
read from db
read from channel
read from channel
read from channel
```

Notes:

* Purposely add a 1 second wait time to the DB reading to simulated slowness.
* The current TTL is 10 seconds
