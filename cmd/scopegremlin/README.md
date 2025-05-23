# scope gremlin

Scope gremlin lists the current discovered scope for each type
definition. The scope considers links to other types, internal and
imported. You can run it without parameters for the current folder.

```
go install github.com/titpetric/exp/cmd/scopegremlin@main
```

In gateway, we get something like:

```yaml
  - name: HostUptimeChecker
    position: host_checker.go:61:6
    references:
    - tunny.Pool
    - sync.Map
    - sync.RWMutex
    - sync.Mutex
    - Gateway
  - name: HostData
    position: host_checker.go:37:6
    references:
    - time.Duration
    - apidef.CheckCommand
```

To list objects by number of couplings, you can use `yq` like so:

```
scopegremlin | yq eval '
  .scopes[] |
  .types | 
  to_entries | 
  .[] | 
  [{"name": .value.name, "references": (.value.references | length ) }] | 
  .[] | "" + .references + " " + .name
' | sort -hr | head -n 10
```

This produces:

```
35 Gateway
19 APISpec
16 URLSpec
9 Test
9 ReverseProxy
9 RedisAnalyticsHandler
6 SessionLimiter
6 JSVM
6 HostCheckerManager
5 WebHookHandler
```

And to do a reverse lookup, how many objects depend on X (uses `yq` and `jq`):

```
scopegremlin | yq eval -o=json '
  .scopes[] |
  .types[] |
  {"name": .name, "references": .references} |
  .references[] as $ref |
  {"type": $ref, "dependents": .name}  
' - | jq -r -s '
  group_by(.type) |
  map({"name": .[0].type, "count": length}) |
  sort_by(.count) |
  reverse |
  .[] | "\(.count) \(.name)"
' | head -n 10
```

This produces:

```
88 null
45 BaseMiddleware
30 Gateway
17 APISpec
11 storage.Handler
11 http.Request
10 logrus.Entry
9 sync.Mutex
8 user.SessionState
8 time.Duration
```

Gateway is referenced 30 times from other types. To list the types:

```
# scopegremlin | yq eval '
  .scopes[] |
  .types[] |
  select(.references[] | contains("Gateway")) |
  .name
' -
```

This gives us a list of 30 declarations referencing Gateway:

```
1. RedisAnalyticsHandler
2. JSVM
3. TykRoundTripper
4. RPCStorageHandler
5. HostCheckerManager
6. HostUptimeChecker
7. DefaultKeyGenerator
8. APIDefinitionLoader
9. DefaultHealthChecker
10. DummyProxyHandler
11. BatchRequestHandler
12. Bundle
13. HTTPDashboardHandler
14. WebHookHandler
15. LogMessageEventHandler
16. JSVMEventHandler
17. ReverseProxy
18. Monitor
19. BaseMiddleware
20. OAuthManager
21. RedisOsinStorageInterface
22. accessTokenGen
23. DefaultSessionManager
24. RedisPurger
25. RedisNotifier
26. Notification
27. ResponseTransformJQMiddleware
28. BaseTykResponseHandler
29. gatewayGetHostDetailsTestCheckFn
30. Test
```