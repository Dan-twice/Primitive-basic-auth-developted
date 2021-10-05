FROM golang:1.16-alpine

WORKDIR /src

# ENV MYSQL_DATABASE recordings

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY *.html ./
COPY *.sql ./
COPY *.pem ./

RUN go build -o /basic-auth-sys
# 8080
EXPOSE 4000  

CMD [ "/basic-auth-sys" ]

# multistaged
# FROM golang:1.16-alpine
# FROM golang:1.16-buster AS build
# FROM mysql

# WORKDIR /app

# ENV MYSQL_ROOT_PASSWORD pR_Mt_93MtMi.
# ENV MYSQL_DATABASE recordings
# ENV MYSQL_USER root
# ENV MYSQL_PASSWORD pR_Mt_93MtMi.

# COPY go.mod ./
# COPY go.sum ./
# RUN go mod download

# COPY *.go ./
# COPY *.html ./

# FROM gcr.io/distroless/base-debian10

# WORKDIR /

# COPY --from=build /web_ser /web_ser

# EXPOSE 8080

# USER nonroot:nonroot

# ENTRYPOINT ["/web_ser"]
# CMD [ "/web_ser" ]