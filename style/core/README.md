# Tyk Go Guidelines Draft

To properly structure code in the go language, you should consider this
as a master guide. It focuses on file layout and source code layout for
packages, while not being opinionated on package structure too much.

## Packages

- Keep lowercase, no underscores, `servicediscovery`.
- Don't shadow packages from stdlib, don't provide your own `http` package, etc.
- Consider internal scope to encourage single responsibility and code reuse

### Documentation

The basic requirement for any package are package docs. This means a
package should provide a `doc.go`, filling out package godocs.
Alternatively, a `README.md` file can be provided in the package and
will be read for documentation.

Ideally, provide godoc and extend it to cover documentation on fields.
There are exceptions (CRUD verb functions), but having complete godoc
coverage is desired.

## Files

Source code files are a representation of the computer program being built, they
are not the program. A simple thought experiment of this is that you can do:

- Merge all files into pkg.go, clean up imports and it will still work.
- Invoke `go build` in package on a single file
- Invoke `go test` in a package on a single file

By doing this, we observe bounded scopes. The file `record.go` may not
need other files in the package (file bound scope), but `record_list.go`
would not pass the build check as it wouldn't know about `Record`.

Small bounded scopes are opportunities for extraction, as those types
that are stand-alone and carry significant complexity, can be extracted
to an internal package. Testing the components inside bound scopes
(packages) is the suggested way to eliminate bugs.

### Import groups

Import blocks are formatted with
[splint fix](https://github.com/titpetric/tools/tree/main/splint), which
rewrites every block to the grouping below in one pass.

The desired grouping is:

~~~go
import (
	// stlib packages

	// 3rd party imports (logrus, assert, etc.)

	// TykTechnologies package imports

	// _ (driver imports)
)
~~~

This is a concern for `task fmt` (`task lint`, `task hooks:pre-commit`).
You shouldn't need to do anything except enforce consistency.

### File declaration order

A file should order declarations from:

- constants
- variables
- types
- functions

This is the desired order.

When defining functions, similar order should be kept:

- arguments
- variables (including return value)
- logic comes after
- don't use named returns, it's mostly not worth it
- named returns + defer are a code smell

It's not desired to have globals, so be mindful when adding new symbols
to existing package scopes. If the package scope is large, the desire is
to decouple it into meaningful components.

- [Uber style guide: Avoid Mutable Globals](https://github.com/uber-go/guide/blob/master/style.md#avoid-mutable-globals)

### File naming restrictions

For a type `T`, a `t.go` is expected, examples:

- ServiceDiscovery, `service_discovery.go`
- Gateway, `gateway.go`
- RateLimitMiddleware, `rate_limit_middleware.go`

A package will commonly violate these, however best practices when not
blocked are to sort the service types into individual files or packages.
With the file naming restrictions, a file gives us a bounded scope for a
service model implementation.

Functions that are attached to this type should live in the same file,
following the desired declaration order. If the functions become too
large, it's possible to split individual ones by file prefix. For
example:

- func (*ServiceDiscovery) Get(), `service_discovery_get.go`.
- func (*HostChecker) Do(), `host_checker_do.go`.

The desired structure is there to ensure that a package protects this
API surface, and ensure that other files in the package don't partition
the implementation. The latter issue is seen in large-scope packages,
which are difficult to navigate.

A package separate to the data model will reference `model.Record`. This
is a good thing for code analysis, as we can track usage between
packages.

Black box tests will similarly reference types and functions from the
package. Static code analysis tells us which functions or models are
referenced from which test. White box tests fail to provide this
usability with basic tools like ctrl+f / `grep` for `model.New`.

### Test file and test function naming restrictions

- Test files should shadow the implementation being tested (`x.go, x_test.go`).
- Test names should match the symbol being tested.
- Black box tests are encouraged for symbol resolution

Template:

- `TestContains()` for testing `Contains()` func
- `TestRecord()` for testing the `Record` type
- `TestRecordGet()` for testing `Record.Get()` func
- `TestGet` may test `Record.Get` or `Get` (max 1 symbol).

To take advantage of grouping, the test can be followed by an `_` and
additional context of what is being tested.

- `BenchmarkGatewayReload_FindGoroutineLeaks`

Bad:

- `TestAllTheRateLimiters`

> As a general guideline, unit tests should be careful to test only the symbol
> referenced, but that doesn't exclude the possibility of indirect testing.
> For example, a `TestAudit` could test any/all functions attached to Audit{},
> demonstrating something similar to an usage example.

- [Uber style guide: Test Tables](https://github.com/uber-go/guide/blob/master/style.md#test-tables)
- [Uber style guide: Avoid unnecessary complexity](https://github.com/uber-go/guide/blob/master/style.md#avoid-unnecessary-complexity-in-table-tests)

### Black box tests

Black box tests use a `_test` suffix for the tested package. It imports
the tested package and provides a clean symbol reference in source.

Black box tests allow you to use only the "public" API parts, so that
internal implementation details are allowed to change without needing to
change any tests. This is the recommended way, otherwise we couple tests
to code, and not to the logic.

It's suggested you write black box tests also because they can be easily
moved to a more appropriate location under `/tests`.

### Integration tests

Integration tests in this context means using an external service for
shared caches or persistent storage (redis, mongo, postgres,...).

The issue with integration tests is that:

- sharing state (redis, postgres, mongo), clearing state from tests
- slow, sequential, unable to skip unit tests vs. integration tests
- running the CI in full against each storage rather than focused

You should pay attention to those restrictions as you write tests. Tests
that clear the database state will interfere with other running tests.
The `/tests` location should be encouraged to segment any integration
tests into their own area to optimize this process in the future.

### General testing

Various micro-optimisations are encouraged:

- use httptest.NewRequest
- use testify/assert (not just standard testing package)
- use t.Cleanup instead of defer in tests

Red flags in testing and product code are usually:

- usage of goroutines without cancellation or lifecycle control
- inline usage of mutexes without a defined API over the data
- usage of atomic.Value or `any` type loosing type safety
- protecting a map with a mutex, but no protection or copies for the stored values
- goroutine leaks, memory leaks, heap escapes (unnecessary copies)

### Type declaration restrictions

For any struct, provide a constructor, returning *T and optionally an
error. As there is some difference in utility, more detailed guide is
provided for the data model, and for the service models.

The constructor gives us an allocation point. Where the allocation needs
to be changed to support new logic, an existing constructor would
minimize the change going forward.

#### Data model restrictions

- constructors generally should not have parameters
- constructors should allocate map values with `make`
- expected form: `func NewT() *T`.
- return an allocated zero value
- no concurrency protections

There is some utility in providing `type Record struct` and `type RecordList
[]Record`, and a common traversal, filtering and retrieval API on
RecordList:

~~~go
type RecordList []*Record
func (RecordList) FilterIDs(ids ...string) RecordList
func (RecordList) Get(id string) *Record
~~~

Similarly, getters should be added to increase the utility of individual
data model fields. This is common practice in gRPC, which generates a getter
for each data model field. Getters usually have a nil check on them, however,
we rarely decouple the data model.

Usage of data model asumes slices and maps should carry a pointer value.

- 1. [pointer values leak less with map[]*T usage](https://github.com/golang/go/issues/20135),
- 2. as it's a pointer, it's directly modifiable in for loops

A simple loop becomes clearer and allows the value being modified in
flight, without having to reference the slice by index.

~~~go
for _, v := recordList {
	v.Name = "John Smith"
}
~~~

Undesired code without pointers:

~~~go
for k, v := recordList {
	v.Name = "John Smith"
	recordList[k] = v
}
~~~

Generally the only requirement for a data model is the struct
definition. Anything else is already evaluation or encoding logic added
to the type. The data model types, for the most part, are meant to go
through a json encoder or decoder, and nothing else.

Repositories or data access objects should be used to provide any
CRUD-like access to data model storage (redis, etc.). A repository can
reference multiple models and combine them to form aggregates (joining
multiple sources of data to form a "view").

- `set` and `get` are two responsibilities.
- `getAndSet` is a race condition.

#### Service model restrictions

- constructors take dependencies like loggers with contextual data,
- constructors should /copy/ inputs and not hold on to references (data sharing/concurrency).
- may start goroutines, use `runtime.SetFinalizer`, take context for cancellation
- don't use embedding for service model, this is not inheritance
- maximum of 1 mutex per struct, concurrency protections

When it comes to the constructors, only a few options are valid:

- `NewT() (*T, error?)`
- `NewT(logger Logger) (*T, error?)`
- `NewT(logger Logger, opts ...OptionT) (*T, error?)`

If a context value is provided to the constructor, we can asume:

- the service object has a lifecycle with cancellation (e.g. run a http server, a polling loop),
- cancelling the context will clean up any goroutines implemented by the service.

Service class objects are meant to be used by the application. For
example, a middleware may use a Cache service and is expected to have
concurrency-safe interface for it.

A service may implement multiple API functions but should be careful to
consider it's responsibility by breaking logic apart in smaller
self-contained functions. A function is self contained if it doesn't use
any other types or symbols other than the receiver, the arguments and
the returned types (no globals, no other package functions, etc.).

A self-contained function is moveable between packages.

### Locality of behaviour

There are various names for this but generally locality of behaviour is
what we're talking about when we mention single responsibility.

An example of shared responsibility is how our storage package has been
implemented. To get a connection, several flags control which connection
you get, essentially making it a complex way to handle what could just
have been three individual connections (or handlers for them).

~~~go
func (rc *ConnectionHandler) getConnection(isCache, isAnalytics bool) model.Connector {
    rc.connectionsMu.RLock()
    defer rc.connectionsMu.RUnlock()
    if isAnalytics {
	return rc.connections[AnalyticsConn]
    } else if isCache {
	return rc.connections[CacheConn]
    }
    return rc.connections[DefaultConn]
}
~~~

Even discounting the unnecessary `else`, a single-connection
handler would be preffered to reason about there only being one
connection within. The accessibility of the connection becomes a
spaghetti code affair, while the constructor looks like this:

~~~go
// NewConnector creates a new storage connection.
func NewConnector(connType string, conf config.Config) (model.Connector, error) {
    cfg := conf.Storage
    if connType == CacheConn && conf.EnableSeperateCacheStore {
	cfg = conf.CacheStorage
    } else if connType == AnalyticsConn && conf.EnableAnalytics && conf.EnableSeperateAnalyticsStore {
	cfg = conf.AnalyticsStorage
    }
    log.Debug("Creating new " + connType + " Storage connection")
~~~

Using these connections is awkward:

~~~go
store := storage.RedisCluster{IsCache: true, ConnectionHandler: gw.StorageConnectionHandler}
store.Connect()
~~~

Obviously the desired code based on this is:

- `func NewConnector(*config.StorageOptionsConf) (model.Connector, error)`.

The responsibility of the constructor is to process the configuration
passed via argument, and return an appropriate error if none is
provided. Ideally the passed arguments never leave the constructor, or
should be passed along as fields.

All of it could be replaced with:

~~~go
return &Gateway{
	CacheStorage:     NewConnector(conf.GetCacheStorage()),
	AnalyticsStorage: NewConnector(conf.GetAnalyticsStorage()),
	Storage:          NewConnector(conf.GetStorage()),
	...
~~~

Here's another example of a standalone implementation of one of those
helpers that provide the config detail:

~~~go
func (c *Config) GetCacheStorage() *conf.StorageOptionsConf {
	if !c.EnableSeparateCacheStore {
		return nil
	}
	return c.CacheStorage
}
~~~

Not only does usage improve, but config evaluation is brought under the
config package umbrella. Reasoning about configuration and how we
evaluate that configuration would improve with the change. The locality
of behaviour is ensured both in the config package, and the gateway
constructor, and shared responsibility is avoided.

### Maintainer

Creator / current maintainer / author: @titpetric.
Reach out for proposals aside the usual dev guides.
