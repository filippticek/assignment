# Build container:
`docker build -t backend-server .`
# Run container:
`docker run -it --rm -p 8080:8080 backend-server:latest`
# Test app in /test folder:
`cd test/`
`go build`
`./test`
# Testing using curl:
## READ id=1
`curl -v -X GET http://localhost:8080/1`
## READ *
`curl -v -X GET http://localhost:8080/1`
## CREATE id=1
`curl -v -X PUT -H "content-type: application/json" http://localhost:8080/ -d '{"Id":1, "status":0, "name":"hp"}'`
## UPDATE id=1
`curl -v -X PUT -H "content-type: application/json" http://localhost:8080/1 -d '{"Id":1, "status":1, "name":"hp"}'`
## DELETE id=1
`curl -v -X DELETE http://localhost:8080/1`
