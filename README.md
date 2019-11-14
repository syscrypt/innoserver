# Innoserver

## Description
The Innoserver is the communication backend for the Innovision project,
an interdisciplinary project of the Aachen University of Applied Science.

## Prerequisites
IMPORTANT!
Never use the default settings in the `Makefile` and the `docker-compose` file
in production. Always remember to set at least the `DBPW` variable in the
`Makefile` and the `MYSQL_PASSWORD` and `MYSQL_ROOT_PASSWORD` variables in the
`docker-compose` file.

## Build
To build the application, simply run the following command:
```sh
make build
```
This triggers the go building process and moves the application's binary
to the `/bin` folder.

TODO: Build and deploy the binary using Docker

## Run
To run the application either use `make run-docker` or `docker-compose up`.
Please notice, that after a fresh building process, the database initialization could take a
while. If you get any errors connecting to the server or don't find your database,
try to wait a little longer.

TODO: Make the application actually do something


## Connect to the database
To connect manually to the `mariadb` server with your mysql client, either use
```sh
make connect
```
or connect to the mysql client in the container
```sh
docker exec -it <container_name> mysql -u root -p
```

## Layout
`./cmd`: Main applications for the project\
`./pkg/repository`: Database interface service definition\
`./pkg/model`: Model definitions for representing datastructures\
`./pkg/handler`: Handler for route administration and handling of requests
