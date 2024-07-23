# 1️⃣🐝🏎️ The One Billion Row Challenge

The One Billion Row Challenge (1BRC) is a fun exploration of how far modern Java can be pushed for aggregating one billion rows from a text file.
Grab all your (virtual) threads, reach out to SIMD, optimize your GC, or pull any other trick, and create the fastest implementation for solving this task!

<img src="1brc.png" alt="1BRC" style="display: block; margin-left: auto; margin-right: auto; margin-bottom:1em; width: 50%;">

The text file contains temperature values for a range of weather stations.
Each row is one measurement in the format `<string: station name>;<double: measurement>`, with the measurement value having exactly one fractional digit.
The following shows ten rows as an example:

```
Hamburg;12.0
Bulawayo;8.9
Palembang;38.8
St. John's;15.2
Cracow;12.6
Bridgetown;26.9
Istanbul;6.2
Roseau;34.4
Conakry;31.2
Istanbul;23.0
```

The task is to write a program which reads the file, calculates the min, mean, and max temperature value per weather station, and emits the results on stdout like this
(i.e. sorted alphabetically by station name, and the result values per station in the format `<min>/<mean>/<max>`, rounded to one fractional digit):

```
{Abha=-23.0/18.0/59.2, Abidjan=-16.2/26.0/67.3, Abéché=-10.0/29.4/69.0, Accra=-10.1/26.4/66.4, Addis Ababa=-23.7/16.0/67.0, Adelaide=-27.8/17.3/58.5, ...}
```

Submit your implementation by Jan 31 2024 and become part of the leaderboard!


## Commands

* `make build` -- Compile the project
* `make run-timed`: -- Execute the 1brc challenge and time it.
* `make clean` -- Clean the project
* `make test` -- Run tests


## License

This code base is available under the Apache License, version 2.

## Results Log

### Baseline: Sequential file read (25s)
See cmd/experiment/fileReadExp.go

This reads the 1b file and times how long it takes. 

### Baseline: mmaped file sequential read (5.6s)
We mmap a file, then read each byte sequentially. 

### 1: Sequential Scanning (146s)
A simple single threaded naive approach of reading the file with a scanner, and then calculating the aggregate measurements. 


### 2: Multi-Threaded publisher and consumers model (8:06)
A go process to read the file sequentially and push lines in to a queue (channel). Then may go consumer processes to read from the queue, and convert the lines into measurements.

**Notes: I'm assuming this was very slow due to the high number of channel reads (for each line). Im assuming under the hoods this requires a mutex acquisition to read from the queue. Therefore, this may have became costly**

### 3: Multi-threaded Consumer Only Model (59s)
We have a main thread which sequentially reads the 10M lines into memory. The we kick of a new go routine to process an array of these lines. The hypothesis is that this model should avoid the locking issue seen in the first model. 

I experimented with different chunk sizes and got the following results on a M1 macbook pro (8 core count)
* 100K = 60s
* 1M = 58s
* 10M = 59s
* 100M = 363s

### 4: Mmaped file with go routine consumers (64s)
We mmpp the data file into memory. We then chunk this data into even chunks for the number of workers we want (aligned on a new line). Each worker processes the chunk in parallel, then we reduce at the end. 

I experimented with the number of workers and got the following results: 
* 1 (basically sequential) = 205s
* 7 = 64s
* 10 = 67s
* 30 = 77s

**I expected this one to be the fastest, and cut significantly on attempt 3. However, it was near the same. My guess is that due to only having 16GB of memory, there is some memory pressure (and page faults) trying to load the entire file into memory at once**

Benchmark Results:
18          68792752 ns/op        83126841 B/op    3002325 allocs/op

### 5: Changed Map to use Measurements instead of measurements pointer (61s)
Our Go worker routines were returning a `map[string]*Measurement`. However, looking at the cpu profile for attempt 4, alot of time is being spent on managing the stack (`runtime.morestack`) and other functions that looked like garbage collection. 

Therefore the idea is by having a `map[string]Measurement`, then some results may live on the stack, and therefore lead to lower garbage collection (and coordination between go routines)

Benchmark results:
19          63,923,818 ns/op        53,225,366 B/op    2,002,570 allocs/op

Time Results:
61.22 real       243.05 user        31.67 sys

### 6: Custom Line Parsing and int64 operations (34s)
Instead of using `strings.Split(line, ";")` and `strconv.ParseFloat(measureString, 64)` we will implement our own string parsing.

Benchmark Results: 
36          33,475,833 ns/op        21,177,892 B/op    1,002,303 allocs/op

Time Results:
34.11 real       124.45 user        20.43 sys

### 7: In place chunk processing (35s)

* Changed lineToMeasure to take a `[]byte`. This allowed the conversion to be more efficient. It also avoids the need to allocate a `string(buffer)` variable which was contributing to a lot of data being placed on the heap

* Changed the chunk parser to parse the city and measurement as we iterate over the line. Before we would read the line until we reach a \n char. Then we would parse the line. Which resulted in use iterating over every character twice. 


Benchmark results:
9         114127301 ns/op        1,012,4064 B/op    1,000,102 allocs/op

Time Results:
35.45 real       130.00 user        22.23 sys

### 8: Changed Map Key 
TODO:
Changed map[string]Measurement to map[unit64]Measurement. This avoids needing to hash the city multiple times for lookup and setting. It also avoids the need to allocate the city name on the heap, which results in less heap space and less GC time