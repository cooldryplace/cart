Cart
====
[Photo](https://www.flickr.com/photos/rmb808/13301196764) [©Roberto Badillo](https://www.flickr.com/photos/rmb808)

## Why
Hey, people want to buy products.

Building this microservice to scale better.

## How
Cut a piece from the monolith and serve chilled.

### Package structure
The package structure is simple for a reason. Currently, this is a straightforward service, so almost everything is in a single package, where business logic, data storage, and API code is located in separate files.
Later if service will become more complex, there will be a need for better granularity and code isolation. But for now, I feel like this is the right balance.

## Limitations
The current DB schema does not allow us to shard data. FKs and autoincrement are in the way.
The suggested next step is to handle constraints in the application code. Start using UUIDs.

