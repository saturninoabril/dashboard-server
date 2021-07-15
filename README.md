# dashboard-server

### Build and Install
Then build and install the binary:

```
$ go install ./...
```

### To spin up the server locally
```
$ make start
```

### Create database migration
Install golang-migrate, refer to https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
```
export DASHBOARD_DATABASE='postgres://dashboarduser:dashboardpwd@localhost:5433/dashboard_dev?sslmode=disable'
migrate create -ext sql -dir store/migrations/migrations_files -seq <change_name>

# Manual migration
migrate -database ${DASHBOARD_DATABASE} -path store/migrations/migrations_files up
migrate -database ${DASHBOARD_DATABASE} -path store/migrations/migrations_files down
```
