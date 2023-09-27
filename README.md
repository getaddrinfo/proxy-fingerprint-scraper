# Fingerprint Scraper

A scraper that loads fingerprints from Discord's `/experiments` endpoint via proxies. Allows for proxy IPs and user agents to be loaded from a file (or a custom source, see `proxy/{ip,ua}/source.go`).

Once fetched, it stores them to a PostgreSQL database, and they can be read via the API.

## Running & Building

To build the program:

```sh
$ git clone https://github.com/getaddrinfo/proxy-fingerprint-scraper.git
$ cd proxy-fingerprint-scraper
$ DB_STRING=... # set an environment variable for convenience

$ GOOSE_DRIVER=postgres GOOSE_DBSTRING=$DB_STRING goose -dir ./migrate up # brings migrations up to date - make sure goose is installed (see https://github.com/pressly/goose)
$ go build -o scraper
$ ./scraper -help # to see arguments
$ DATABASE_URL=$DB_STRING ./scraper -fingerprints -workers 10
```

## Documentation

Some simple documentation regarding the API is accessible under `/` - to access it, follow the steps below:

```sh
$ psql -h localhost -p 5432 -U <user>
# ...
# \c <db>
# UPDATE auth SET token = '...' WHERE id = 0; -- set your token to a value that is 32 characters long
$ curl http://localhost:48832/?token=...
```

## TODO

- [ ] Docker containers
- [ ] Better Admin UX