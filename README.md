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