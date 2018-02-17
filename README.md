go-yahtzee
==========

A Go implementation of an optimal Yahtzee player on the Raspberry Pi. See more about this project [on my blog](https://tim.palpant.us), or [play along online](https://tim.palpant.us/yahtzee).

[![GoDoc](https://godoc.org/github.com/timpalpant/yahtzee?status.svg)](http://godoc.org/github.com/timpalpant/yahtzee)
[![Build Status](https://travis-ci.org/timpalpant/yahtzee.svg?branch=master)](https://travis-ci.org/timpalpant/yahtzee)
[![Coverage Status](https://coveralls.io/repos/timpalpant/yahtzee/badge.svg?branch=master&service=github)](https://coveralls.io/github/timpalpant/yahtzee?branch=master)

Installation
============

Score tables
------------

You will need to build the score tables first, using the `compute_scores` tool:

```
$ go get github.com/timpalpant/yahtzee
$ cd cmd/compute_scores
$ go build
```

Build the expected value score tables:

```
$ ./compute_scores -logtostderr -observable expected_value -output expected-value.gob.gz
```

Build the high score tables:

```
$ ./compute_scores -logtostderr -observable score_distribution -output score-distribution.gob.gz
```

The expected value tables are 5.7 MB and the high score tables are 1.8 GB on disk.

Web server
----------

Build and run the web server, passing the location of the score data tables:

```
$ go install github.com/timpalpant/yahtzee/cmd/yahtzee_server
$ yahtzee_server -logtostderr -port 8080 -expected_scores expected-scores.gob.gz -score_distributions score-distributions.gob.gz
```

The server will take a few minutes (and GB of RAM) to load the score tables at startup. Navigate to http://localhost:8080.

Image processing server
-----------------------

The image processing web service is written in Python. The easiest way to build and run it is with Docker:

```
$ cd image_processor
$ docker build -t yahtzee-image-processor:latest .
$ docker run -p 8080:8080 yahtzee-image-processor:latest
```

A pre-built image is also available at: `docker.io/timpalpant/yahtzee-image-processor:latest`.

Raspberry Pi Player
-------------------

First install GoCV, as described here: https://gocv.io/getting-started/

Then build the RPi player with:

```
$ cd cmd/rpi
$ go build
```

and run with:

```
$ ./rpi -logtostderr -v 3 -yahtzee_uri http://localhost:8085 -image_processing_uri http://localhost.local:8080 -score_to_beat 300
```

It is recommended to run the yahtzee server and image processing service on a separate machine,
since they require more memory and computational resources than available on the pi.

## License

This package is released under the [GNU Lesser General Public License, Version 3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html)

Yahtzee is a registered trademark owned by Hasbro. Hasbro is not affiliated with this project.
