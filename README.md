Cart
====
![Photo](https://live.staticflickr.com/3687/13301196764_dd38b5a7e3_n.jpg)

[©Roberto Badillo](https://www.flickr.com/photos/rmb808)

[![Go Report Card](https://goreportcard.com/badge/github.com/cooldryplace/cart)](https://goreportcard.com/report/github.com/cooldryplace/cart)
[![GoDoc](https://godoc.org/github.com/cooldryplace/cart?status.svg)](https://godoc.org/github.com/cooldryplace/cart)

* [gRPC sources](https://github.com/cooldryplace/proto-sources)
* [generated gRPC code](https://github.com/cooldryplace/proto)
* [TLS certificates](https://github.com/cooldryplace/certs)

## Why
Hey, people want to buy products.

Building this microservice to scale better.

## How
Cut a piece from the monolith and serve chilled.

### API documentation
Have to be generated. So I have generated a client with GoDoc.
[Client documentation](https://godoc.org/github.com/cooldryplace/proto#CartsClient)

### Package structure
The package structure is simple for a reason. Currently, this is a straightforward service, so almost everything is in a single package, where business logic, data storage, and API code is located in separate files.
Later if service will become more complex, there will be a need for better granularity and code isolation. But for now, I feel like this is the right balance.

#### Future structure
In the future packages structure will be derived from domains, and not from the functionality that code provides.
An example would be:
* user
* billing
* auth

Not:
* controller
* service
* dao

[@rakyll Style guideline for Go packages](https://rakyll.org/style-packages/)

### Other microservices
Carts rely on other microservices.
* AuthZ: to check that the caller is allowed to access/modify Cart state.
* Billing: to get prices.
* Business Analytics: to submit Cart updates details and make conclusions.
* Products: to get product details.

Some microservices rely on Cart.
* Orders: to retrieve Cart state.
* Users: to retrieve Cart(s) for User profile.

### Testing
To run all tests: `go test github.com/cooldryplace/cart/...`.

To skip integration tests: `go test --short github.com/cooldryplace/cart/...`.

#### Missing Parts
The current integration test does not cover cases when storage or other dependency fails.

### Metrics
Currently, only API and client metrics are implemented. This already allows measuring availability and latency. In the future, I can add more specific metrics that will not be used in SLO implementations but will be in dashboards. To pinpoint the root cause of a problem during incidents.
So we alert based on SLIs, look at the dashboard, and know where and why it happens.

### Tracing
Adding spans here and there will help to identify bottlenecks. I use Opencensus with Stackdriver exporter for this.

## Limitations
The current DB schema does not allow us to shard data. FKs and autoincrement are in the way.
The suggested next step is to handle constraints in the application code. Start using UUIDs.

