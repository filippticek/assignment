First build container:

docker build -t backend-server .

Run container:

docker run -it --rm -p 8080:8080 backend-server:latest

Test app in /test folder:

cd test/

go build 

./test
