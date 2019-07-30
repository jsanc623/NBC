# Implementing CitiBike's Stations API using Go

A RESTful API using Go which queries and caches CitiBike's Stations API


## Behind the Scenes

This project utilizes one of my personal projects, 
[Sphire Mantis](https://github.com/sphireco/mantis). It's a small library
which provides interfaces for logging, error handling, URL query/param fetching
(using Mux), as well as a small interface for MySQL. 

Additionally, it makes use of Allegro BigCache for caching the CitiBike data, 
as well as VictorSpringer's http-cache as an in memory HTTP middleware. Lastly, it uses
Subosito's gotenv to load configuration data from a .env file, as well 
as Gorilla Mux. 

[Allegro BigCache](github.com/allegro/bigcache)<br/>
[Sphire Mantis](github.com/sphireco/mantis)<br/>
[Subosito Gotenv](github.com/subosito/gotenv)<br/>
[Gorilla Mux](github.com/gorilla/mux)<br/>
[VictorSpringer http-cache](github.com/victorspringer/http-cache)

## Build, Test, Run

To build this project:
```bash
git clone github.com/jsanc623/NBC
cd NBC
go build
```

To run:
```bash
NBC.exe  (or ./NBC if on *nix)
```

To test:
```bash
go test
```

## The API

The project will expose it's APIs on http://127.0.0.1:4000, though this can be changed
by editing `SRV_ADDRESS` and `SRV_PORT` in the .env file

Where noted, `[paged]` corresponds to an endpoints ability to be paginized
using `?page=n` where `n` is the page number. 

By default, it will return 20 items per page. However, if `[limited]` is present,
it will take a limiter per page using `?perPage=n` where `n` is the number of items per page.
 
When both appear, such as `[paged, limited]`, they can be used in conjunction e.g. `?page=2&perPage=5`

##### /stations `[paged, limited]`
`GET` Gets all stations

##### /stations/in-service `[paged, limited]`
`GET` Gets all stations that are in service

##### /stations/not-in-service `[paged, limited]`
`GET` Gets all stations that are not in service

##### /stations/:searchString `[paged, limited]`
`GET` Performs a case-insensitive search of :searchString on all
stations and returns those which have a match in either the name or 
address.

##### /dockable/:stationId/:bikesToReturn
`GET` Returns a boolean and message which denote whether there are
enough docks available at the given :stationId to fit the number of :bikesToReturn


## Caching

I utilized [http-cache](github.com/victorspringer/http-cache), which 
adds a middleware around each request. The cache eviction policy is 
LRU. Initially I had implemented [Bluele's GCache](https://github.com/bluele/gcache), 
however I would have had to write my own middleware to wrap the HTTP requests. I chose 
this initially because it provides an ARC eviction policy (mixture of LRU and LFU)

Additionally, the CitiBike JSON is cached using [Allegro BigCache](github.com/allegro/bigcache), which
is a fast and efficient in memory cache. Using this dropped initial loads from 700ms to 800ms to
just over 300ms (once `http-cache` is warmed up for an endpoint, it's typical to see 10ms to 16ms response times)

## Logging and Error Handling

I used Mantis for logging and error handling, which is my own personal project, and which I made public 
for the purposes of this exercise. I have not written tests for it yet. 

It provides a wrapper around `log` and exposes it to main.Logger, from which
it can be used throughout the application. Additionally, each request is 
captured by a custom middleware implemented in this project which
utilizes Logger as its interface. 

## Routing

All routes are defined in `routes.json` and are loaded when the application is 
initiated. Ideally, `reflect` would be used to load the handlers defined
in `routes.json`, however I chose to just use a switch and assign. 