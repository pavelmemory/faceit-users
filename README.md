## Test task: web-service to CRUD users

### Local run

Please make sure you have Docker installed on your system.
```bash
docker -v
```

The best way to start testing of the service locally is to use `docker-compose`.    
Please make sure the port `8080` is not allocated by any other service/daemon/application before running commands listed below.  
To start the service and its dependencies you need to run:
```bash
make integration-env-up
```
it will:
1. create container of the PostgreSQL database and setup fresh schema in it.
1. create image of the `faceit-users` service and start it.

The service should be ready for use after a couple of seconds. You can check if it is ready by running:
```bash
make integration-env-ready
```

Once the service is ready you could try to check some basic info about it:
```bash
curl localhost:8080/-/version
```

To create a user please run:
```bash
curl -v -H 'Content-type: application/json' \
    -d '{"first_name": "fn", "last_name":"ln", "nickname":"nn", "email":"ue@mail.com", "password": "password", "country":"XX"}' \
    localhost:8080/users
```

To get a user:
```bash
curl -v localhost:8080/<Location>
```
where <Location> is the value returned in `Location` header of the previous operation result without leading slash.

To change a user:
```bash
curl -v -H 'Content-type: application/json' \
    -X PUT \
     -d '{"first_name": "fnn", "last_name":"lnn", "nickname":"nnn", "email":"uee@mail.com", "country":"XY"}' \
    localhost:8080/<Location>
```

And finally to remove the user:
```bash
curl -v -X DELETE localhost:8080/<Location>
```

_TODO:_ List users based on the filtration request.

The flow described above is also available as an integration test that could be run by the command:
```bash
make integration-test
```
**NOTE**: it may fail if user is already present in the database.

### Not covered:

- listing of user entities with filtering
- no notifications send on the user update events
- no automatic migration of database schema
- no metrics exported
- no proper README.md file with listing of configuration settings supported
- no OpenAPI specification of the endpoints
- the lack of test for functionality (especially for the `storage` package)
- caching of the user information to reduce the load on the database
- authorization and authentication of incoming requests
- support of the feature flags
- dynamic change of the logging level for certain endpoints to get better visibility in usgent cases
- client lib for the service that could improve integration with it
- hardcoded configuration values in the code
- ... etc.