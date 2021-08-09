<h1 align="center">Welcome to Hyperwork-Backend 👋</h1>
<p>
  <img alt="Version" src="https://img.shields.io/badge/version-1.0.0-blue.svg?cacheSeconds=2592000" />
</p>

> This project is intended to support companies in the implementation of hybrid working.

## Install

```sh
go build timecool/hyperwork
```

## Start Project

```sh
go run timecool/hyperwork
```

When the first user is created, you must first approve the member or administrator.
```json
db.users.updateOne({"_id" : "UUID"},{$set: { "role" : "admim"}});
```
## Start Tests
a database connection is needed for the tests
```sh
go test ./... 
```


## Author

👤 **Vincent Bärtsch**

* Github: [@timecool](https://github.com/timecool)