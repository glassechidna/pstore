# `pstore`

[![Build Status](https://travis-ci.org/glassechidna/pstore.svg?branch=master)](https://travis-ci.org/glassechidna/pstore)

`pstore` is a tiny utility to make usage of [AWS Parameter Store][aws-pstore] an
absolute breeze. Simply prefix your application launch with `pstore exec <yourapp>`
and you're up and running - in dev or prod.

[aws-pstore]: https://aws.amazon.com/ec2/systems-manager/parameter-store/

## Usage

`pstore` expects the `AWS_REGION` environment variable to be set to the region
that your parameters are stored in.

### `exec`

```
AWS_REGION=us-east-1 PSTORE_DBSTRING=MyDatabaseString pstore exec -- 'echo val is $DBSTRING'
val is SomeSuperSecretDbString
```

`pstore` is usable out of the box. By default it looks for environment variables
with a `PSTORE_` prefix. For example, `PSTORE_DBSTRING=MyDatabaseString` asks
AWS to decrypt the parameter named **MyDatabaseString** and stores the decrypted
value in a new environment variable named `DBSTRING`. If there are no envvars
with the `PSTORE_` prefix, it's essentially a noop - so the same command can be
used in local dev and in prod.

If `pstore` fails to decrypt any envvars it will exit instead of launching your
application.

### `shell`

Sometimes you don't want to exec the child process directly. You want to use the decrypted values as part of a larger script. In that case you can do:

```
#!/bin/bash
# do some stuff ...
eval $(PSTORE_DBSTRING=MyDatabaseString pstore shell)
echo $DBSTRING # will echo out your secret string!
```

### `powershell`

Same as the above, albeit for our Windows friends.

```
$Env:PSTORE_DBSTRING = "MyDatabaseString"
$Cmd = (pstore powershell mycompany-prod) | Out-String
Invoke-Expression $Cmd
Do-SomethingWith -DbString $DBSTRING
```



## Advanced

`pstore` also works with tagged parameters, which can be helpful when you have
a _lot_ of parameters and don't want to enumerate them all individually. You can
specify `PSTORETAG_tagkey=tagval` and `pstore` will retrieve all parameters with
`tagkey=tagval`. `pstore` will expect to find an additional tag on these parameters,
`pstore:name=ENVVAR`. `pstore` then sets `ENVVAR=value` in the environment.

The `PSTORE_` and `PSTORETAG_` prefixes are configurable if you want to use 
something else. If you want to use `MYSECRETS_` as a prefix, simply invoke
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
RUN curl -sL -o /usr/bin/pstore https://github.com/glassechidna/pstore/releases/download/1.2.0/pstore_linux_amd64
RUN chmod 0755 /usr/bin/pstore
ENTRYPOINT ["pstore", "exec", "--verbose", "--"]
CMD env
```
