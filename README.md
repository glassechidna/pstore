# `pstore`

[![Build Status](https://travis-ci.org/glassechidna/pstore.svg?branch=master)](https://travis-ci.org/glassechidna/pstore)

`pstore` is a tiny utility to make usage of [AWS Parameter Store][aws-pstore] an
absolute breeze. Simply prefix your application launch with `pstore exec <yourapp>`
and you're up and running - in dev or prod.

[aws-pstore]: https://aws.amazon.com/ec2/systems-manager/parameter-store/

## Usage

`pstore` is usable out of the box. By default it looks for environment variables
with a `PSTORE_` prefix. For example, `PSTORE_DBSTRING=MyDatabaseString` asks
AWS to decrypt the parameter named **MyDatabaseString** and stores the decrypted
value in a new environment variable named `DBSTRING`. If there are no envvars
with the `PSTORE_` prefix, it's essentially a noop - so the same command can be
used in local dev and in prod.

If `pstore` fails to decrypt any envvars it will exit instead of launching your
application.

The `PSTORE_` prefix is configurable if you want to use something else. If you
want to use `MYSECRETS_` as a prefix, simply invoke
`pstore exec --prefix MYSECRETS_ <yourapp>`.

Finally, for debugging there is the `pstore exec --verbose <yourapp>` flag.
Before launching, `pstore` will output what its doing to stdout, e.g.

```
$ pstore exec --verbose <yourapp>
✔ Decrypted MYREALSECRET︎
✗ Failed to decrypt PstoreVal (MYLAMESECRET)
ERROR: Failed to decrypt some secret values
```

## Docker

`pstore` is well-suited to acting as an entrypoint for a Dockerised application.
Adding it to your project is as simple as:

```
FROM alpine
RUN apk add --update curl
RUN curl -sL -o /usr/bin/pstore https://github.com/glassechidna/pstore/releases/download/1.0.3/pstore_linux_amd64
RUN chmod 0755 /usr/bin/pstore
ENTRYPOINT ["pstore", "exec", "--verbose", "--"]
CMD env
```
