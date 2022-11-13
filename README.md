# Innoserver (Archived, Project 2019)

## Description

The Innoserver is the communication backend for the Innovision project,
an interdisciplinary project of the Aachen University of Applied Science.
Most functionalities can be executed through the `Makefile` targets listed
below.

## Prerequisites

### Required Technologies:

To build and run the full system, you will need the following software:

- `make`
- A running `docker` engine
- `docker-compose`
- `go toolchain`
- `swagger` https://github.com/go-swagger/go-swagger/releases/tag/latest
- A `mysql` <b>client</b> software

### IMPORTANT!

Never use the default settings in the `Makefile` and the `docker-compose` file
in production. Always remember to set at least the `DBPW` variable in the
`Makefile` and the `MYSQL_PASSWORD` and `MYSQL_ROOT_PASSWORD` variables in the
`docker-compose` file.

## Makefile Targets

### Hints

If you want to build and run the application for the first time, I recommend to
take a look at the following remarks:

- The database container needs <b>a lot of</b> time to initialize the database.
  So be patient and plan in some wait time in advance (ca. 10 - 15min). Don't panic,
  this has only to be done once at creation. Afterwards the db-server just has to be
  started.
- After the database container finished the initialization process you should
  import the database schema file and the demodata provided in the `/init` folder.
  For that you have to execute first the `init-database` and then the
  `demodata` target.
- In order to test the latest api in the `swagger-ui`, first execute the
  `swag-gen-doc` target and then the `swagger-ui` target.
- If you built and run the application with the `all` target you can simply
  shutdown the whole system by executing the `shutdown` target.
- With `make connect` you can directly access the innovision database
  on the mariadb server (if you have a `mysql` client software installed)

### Build and run application all at once

Use this target if you want to build and run everything (DB-Server, Application,
Swagger-Ui).

With a simple `make`, a `swagger` documentation is created through the in-code
annotations under the `swag-gen-doc` target. Then the `build` target is executed
which creates the application's binary. After that the database server is started.
Finally the application starts and connects to the db-server, processing incoming
requests. Make sure your database server is fully initialized before trying to
run the application.

## Build the application

To build the application, simply run the following command:

```sh
make build
```

This triggers the `go` building process and moves the application's binary
to the `/bin` folder.

TODO: Build and deploy the binary using Docker

## Run

To run the application either use `make run-docker` or `docker-compose up` in
combination with the `run` target.

TODO: Run the application using the compose file inside a container.

## Connect to the database

To connect manually to the `mariadb` server with your mysql client, either use

```sh
make connect
```

or connect to the mysql client in the container

```sh
docker exec -it <container_name> mysql -u root -p
```

## Swagger Documentation

To generate (and validate) the `Swagger` Documentation, execute the `swag-gen-doc` target.
With the `swag-doc` target you can serve the documentation server to take a look
on the API-Specification.

## Swagger UI

To host the Swagger-ui and test the API functions, execute the `swagger-ui` target. This
will build and run a docker container providing the ui on the specified port `SWAGGERUIPORT`
in the Makefile (default `9000`).

## Shutdown the whole system

The `shutdown` target kills all processes circumstancing the application. Be careful, since
this command kills all programms over it's name.

## Layout

`./cmd`: Main applications for the project\
`./pkg/repository`: Database interface service definition\
`./pkg/model`: Model definitions for representing datastructures\
`./pkg/handler`: Handler for route administration and handling of requests
