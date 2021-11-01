# kusk-gateway
Kusk-gateway is the API Gateway, based on Envoy and using OpenAPI specification as the source of configuration

## Steps to setup local development cluster and deploy kusk-gateway operator

#### Create local cluster
The local cluster setup depends on having a k3d registry named reg available
- `make create-env`

#### Delete local cluster
- `make delete-env`

#### Deploy an example API
`kubectl apply -f examples/httpbin && kubectl rollout status -w deployment/httpbin`
This will create a deployment and load balancer service for httpbin.
The openapi / swagger document describing the API will be applied in the form of a kusk API CRD.
This will cause the reconcile loop in the kusk-gateway control plane to kick in and update the envoy config to allow you 
to curl the service.

Envoy is listening on port 8080, the IP will be published after `create-env` succeeds.

The httpbin service in this example is available at the basepath `/` so to call the get endpoint on the httpbin service 
we simply run `curl http://<published-IP>:8080/get`

# Local development with docker-compose

This development mode utilises an ability of kusk-gateway to consume OpenAPI file directly.

Preliminary steps:

```shell
# From the project root
cp development/.env.example ./.env
```

.env file has variables that control Docker stack behaviour.

For the development change PROJECT_ROOT/.env file to point GO_CONTROL_PLANE_ADDRESS and GO_CONTROL_PLANE_PORT variables to ip address and port your kusk-gateway is listening on.
This will allow Envoy instance to connect to your application in IDE.

Front-envoy will generate configuration from envoy.yaml.tmpl with "default" Envoy cluster name and Node ID based on ENVOY_CLUSTER + HOSTNAME.

Kusk-gateway will consume OpenAPI file, passed with "--in" parameter and will switch to "local" mode that will skip Kubernetes connection.

Once Front Envoy starts, it will connect to kusk-gateway with GRPC with its NodeID and Cluster ("default") fields specified and kusk-gateway will provide generated configuration.

To run with kusk-gateway being developed in IDE:

```shell
# From the project root
# Make sure .env has GO_CONTROL_PLANE_ADDRESS=<IP_ADDRESS_OF_APP_IN_IDE>
docker-compose up
```

To run with kusk-gateway built as Docker container:

```shell
# From the project root
# Make sure .env has GO_CONTROL_PLANE_ADDRESS=kusk-gateway before running this.
docker-compose --profile gateway up
```

To run with kusk-gateway and mock server:

```shell
# From the project root
# Make sure .env has GO_CONTROL_PLANE_ADDRESS=kusk-gateway before running this.
docker-compose --profile gateway --profile mock up
```

By default kusk-gateway in Docker mode uses ./development/petshop-openapi-short-with-kusk-and-mock.yaml file with mocking enabled on some endpoints.

Envoy frontends will be available on *http://172.21.0.5:8080* (Cluster1) and *http://172.21.0.6:8080* (Cluster2) while backend (petstore app) could be reached on http://172.21.0.3:8080 .

On MacOS, the frontends are available on *http://localhost:8080* (Cluster1) and *http://localhost:8081* (Cluster2)

Envoy management interface is available on *http://172.21.0.5:19000*,  *http://172.21.0.6:19000*, there one can verify what configuration it has in config_dump.

On MacOS, the Envoy management interface is available on *http://localhost:19000* and *http://localhost:19001*  

Mock server will be available on *http://172.21.0.10:8080*

On MacOS, Mock server will be available on *http://127.0.0.1:8082*

To test (depends on configured variables in your OpenAPI file):

```shell
# Linux
curl -v -X GET 'http://172.21.0.5:8080/petstore/api/v3/pet/findByStatus?status=available' -H 'accept: application/json'

# MacOS
curl -v -X GET 'http://localhost:8080/petstore/api/v3/pet/findByStatus?status=available' -H 'accept: application/json'
```

For the convenience you can use provided petshop-openapi-short-with-kusk.yaml or petshop-openapi-short-with-kusk-and-mock.yaml files in ./development.
