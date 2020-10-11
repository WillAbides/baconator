# baconator

[![godoc](https://godoc.org/github.com/WillAbides/baconator?status.svg)](https://godoc.org/github.com/WillAbides/baconator)
[![ci](https://github.com/WillAbides/baconator/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/baconator/actions?query=workflow%3Aci+branch%3Amaster+event%3Apush)

Baconator is a json api to find links between movie actors.  It's inspired by 
[The Oracle Of Bacon](https://www.oracleofbacon.org/) and is a learning 
exercise for me.

## Installation

`go get github.com/willabides/baconator/cmd/baconator`

## Usage

`baconator -l <tcp address> -data <path to data.tar.bz2>`

If the data file doesn't already exist at the given path, baconator will 
download it for you.

## API

Baconator currently requires that actor names be spelled exactly like their 
wiki page.

There are two endpoints:

### `/link?a=:actor&b=:actor`

This returns the link between two actors.

```
$ curl -s "http://localhost:8239/link?a=James+Dean&b=Kevin+Bacon" | jq .
[
  {
    "name": "James Dean",
    "type": "cast"
  },
  {
    "name": "East of Eden (film)",
    "type": "movie"
  },
  {
    "name": "Julie Harris",
    "type": "cast"
  },
  {
    "name": "The Split (film)",
    "type": "movie"
  },
  {
    "name": "Donald Sutherland",
    "type": "cast"
  },
  {
    "name": "Animal House",
    "type": "movie"
  },
  {
    "name": "Kevin Bacon",
    "type": "cast"
  }
]
```

### `/center?p=:actor`

This returns information that would be found on Oracle of Bacon's 
[onecenter page](https://www.oracleofbacon.org/onecenter.php)

```
$ curl -s "http://localhost:8239/center?p=Kevin+Bacon" | jq .
{
  "count_by_distance": {
    "0": 1,
    "1": 857,
    "10": 3,
    "2": 61663,
    "3": 224174,
    "4": 96516,
    "5": 6495,
    "6": 496,
    "7": 50,
    "8": 12,
    "9": 15
  },
  "total_linkable": 405043,
  "average_distance": 3.009139770345371
}
```
