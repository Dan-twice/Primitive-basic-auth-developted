# First stage — build
# image
FROM golang:1.16-alpine as builder

# doesn't exit will be created
WORKDIR /intermediary_container

# copy to vitual linux machine
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY /assets ./
COPY *.go ./
# COPY /scripts ./

# build the execuatable
RUN go build -o /app

# # Second stage — sized
# # production image, only sample of exec files, small
# # Environment + Compiled output = Artifact
FROM alpine:latest

# # copy binary executable only
COPY --from=builder /app /app
COPY /assets .
COPY /migr .

EXPOSE 4000  

CMD [ "/app" ]
