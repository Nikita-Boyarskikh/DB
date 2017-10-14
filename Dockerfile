FROM ubuntu:16.04
MAINTAINER Nikita-Boyarskikh <N02@yandex.ru>

# Install base packages
RUN apt update &&\
    apt-get install -y software-properties-common python-software-properties &&\
    add-apt-repository ppa:gophers/archive &&\
    apt update &&\
    apt-get install -y postgresql golang-1.8-go git

USER postgres

# Work around postgresql
RUN service postgresql start &&\
    psql -c "CREATE ROLE forums WITH LOGIN ENCRYPTED PASSWORD 'forums_admin_pass'" &&\
    psql -c "CREATE DATABASE forums_db;" &&\
    psql -c "GRANT ALL ON DATABASE forums_db TO forums;" &&\
    psql -d forums_db -c "CREATE EXTENSION IF NOT EXISTS citext;" &&\
    service postgresql stop

USER root

# Set environments
ENV REPO=github.com/Nikita-Boyarskikh/DB \
    PGPASSWORD=forums_admin_pass PGUSER=forums PGDATABASE=forums_db PGHOST=127.0.0.1 PGPORT=5432 \
    GOROOT=/usr/lib/go-1.8 GOPATH=/opt/go
ENV PATH="$GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH"

# Get update
RUN go get -u github.com/mailru/easyjson/...\
              github.com/jackc/pgx\
              github.com/pkg/errors\
              github.com/valyala/fasthttp\
              github.com/qiangxue/fasthttp-routing\
              github.com/op/go-logging

# Copy sources into container
RUN mkdir -p $GOPATH/src/$REPO
COPY . $GOPATH/src/$REPO/
WORKDIR $GOPATH/src/$REPO

# Create database structure
RUN service postgresql start &&\
    psql -f assets/sql/create.sql &&\
    service postgresql stop

# Migrate and generate
RUN for i in $(find assets/sql/migrations -name '*.sql'); do psql -f $i; done && bin/gen

# Start
EXPOSE 5000
CMD service postgresql start && go run main.go --conf=config/config.json