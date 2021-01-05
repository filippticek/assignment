FROM golang:1.15-alpine
RUN apk add sqlite
RUN apk add gcc
RUN apk add musl-dev
RUN mkdir /backend
COPY ./backend/ /backend
WORKDIR /backend
RUN go build -o server .
CMD [ "/backend/server" ]
