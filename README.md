# sia-antfarm

sia-antfarm is a collection of utilities for performing complex, end-to-end
tests of the [Sia](https://github.com/NebulousLabs/Sia) platform.  These test
are long-running and offer superior insight into high-level
network behaviour than Sia's existing automated testing suite.

# Prerequisites

You must have `siad` installed.  If it's outside of your path, provide its location using the `-siad` flag.  For all flags, see `sia-antfarm -h`.

# Install

`make`

# Running a sia-antfarm

This repository contains two utilities, `sia-ant` and `sia-antfarm`.  `sia-ant` starts up a `siad` instance and runs jobs using that instance.  Jobs can be toggled on using flags, see `sia-ant -h` for a list of jobs.  `sia-antfarm` creates a number of `sia-ants`, running a configurable array of jobs.  `sia-antfarm` takes one flag, `-config`, which is a path to a JSON file defining the number of ants and the job(s) for each ant to perform.  For example,


`config.json:`
```json
[
	{
		"APIAddr": "localhost:9980",
		"RPCAddr": "localhost:9981",
		"HostAddr": "localhost:9982",
		"Jobs": [
			"gateway"
		]
	},
	{
		"Jobs": [
			"gateway"
		]
	},
	{
		"Jobs": [
			"gateway"
		]
	},
	{
		"Jobs": [
			"gateway"
		]
	},
	{
		"Jobs": [
			"gateway",
			"mining"
		]
	}
]
```

This `config.json` creates 5 `sia-ants`, with four running the `gateway` job
and one running a `gateway` and a `mining` job.  If `HostAddr`, `APIAddr`, or
`RPCAddr` are not specified, they will be set to a random port. 


# License

The MIT License (MIT)
