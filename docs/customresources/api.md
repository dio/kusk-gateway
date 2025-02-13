# API

This resource uses OpenAPI file with x-kusk annotations as the source of truth to configure routing.
Refer to [extention](../extension.md) for the further information on how to add routing information to OpenAPI file.

The required field of API resource is spec.**spec** where `x-kusk`-enhanced OpenAPI file is supplied as an embedded string. You can generate it (and integrate into your CI) using [kgw](https://github.com/kubeshop/kgw) CLI tool.

The optional spec.**fleet** field specifies to what Envoy Fleet (Envoy Proxy instances with the exposing K8s Service) this configuration applies to.
fleet.**name** and fleet.**namespace** reference the deployed EnvoyFleet Custom Resource name and namespace.
You can deploy you API configuration in any namespace with any name and it will be applied to the specific Envoy Fleet.
If this option is missing, the autodetection will be performed to find the single deployed in the Kubernetes cluster fleet which is thus considered as the default fleet.
The deployed API custom resource will be changed to map to that fleet accordingly.
If there are multiple fleets deployed, the spec.**fleet** is required to specify in the manifest.

Once the resource manifest is deployed, Kusk Gateway Manager will use it to configure routing for Envoy Fleet.
Multiple resources can exist in different namespaces, all of them will be evaluated and the configuration merged on any update with these resource.
Trying to apply a resource that has conflicting routes with the existing resources (i.e. same HTTP method and path) would result in error.

**Alpha Limitations**:

* currently resource **status** field is not updated by manager when the configuration process finishes.

*Example*

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: api-sample
spec:
  # Envoy Fleet (its name and namespace) to deploy the configuration to, here - deployed EnvoyFleet with the name "default" in the namespace "default".
  # Optional, if not specified - single (default) fleet autodetection will be performed in the cluster.
  fleet:
    name: default
    namespace: default
  # OpenAPI file with x-kusk annotation here
  spec: |
    openapi: 3.0.2
    servers:
      - url: /api/v3
    info:
      description: Some description
      version: 1.0.0
      title: the best API in the world
    # top level x-kusk extension to configure routes
    x-kusk:
      disabled: false
      hosts: [ "*" ]
      cors:
        origins:
        - "*"
        methods:
        - POST
        - GET
        - OPTIONS
        headers:
        - Content-Type
        credentials: true
        expose_headers:
        - X-Custom-Header1
        max_age: 86200
      upstream:
        service:
          name: oldapi
          namespace: default
          port: 80
        # Strips prefix when forwarding to upstream
        rewrite:
          pattern: "^/api"
          substitution: ""
      path:
        prefix: /api
    paths:
      /pet:
        x-kusk:
          disabled: true
        post:
          x-kusk:
            disabled: false
            upstream:
              host:
                hostname: newapi.default.svc.cluster.local
                port: 8080
    --- skipped ---
        put:
          summary: Update pet
          description: Update an existing pet by Id
          operationId: updatePet
     --- skipped ---

```
