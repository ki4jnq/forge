# Forge
#### Forge is an Ops/Automation tool that aims to be platform agnostic

## Installation

The easy way:
```bash
curl "https://forge-dist.s3.amazonaws.com/install-forge" | bash -s
```

This will put the `forge` command into your `/usr/local/bin` and should be available immediately.

To build from source, you must setup a Go development environment and Glide. Assuming you've already
setup Go and Glide, you should be able to use the following commands to build and install Forge on your
system.

```bash
cd $GOPATH/src/github.com/ki4jnq/ # Create this directory if it does not exist.
git clone git@github.com:ki4jnq/forge.git
cd forge
glide up
go install ./cmd/forge
```

After running these commands, you should be able to execute `forge` from a command line (provided you have added
`$GOPATH/bin` to your `PATH`).

If you do not have Go or Glide use the following instructions:

- Go: https://golang.org/doc/install
- Glide: https://github.com/Masterminds/glide#install (hint: Use `brew` on OSX)

Eventually, Forge will have distributed binaries that can be downloaded directly to your computer.

## Usage and Commands

Forge follows a typical git/aws-cli command structure, so all commands are in the form of:

```bash
forge SUBCOMMAND [--env environment] [opts]
```

Currently, the following subcommands are available:
1. deploy
2. db
3. run

### Configuration

Forge is configured via a YAML file named `Forgefile` which should be placed at the root level of your project. The Forgefile
is processed through the Golang [Text Template Engine](https://golang.org/pkg/text/template/) before being parsed, which allows
you to create dynamic configs that read ENV vars.

All top-level keys in the Forgefile correspond to an environment, such as `development` or `production`. Within these
top-level environments are configurations for individual sub-commands, e.g. `deploy`. All further levels are specific to the
individual sub-commands.

  An example configuration could look something like this:

```yaml
development:
  run:
    './my-server':
    - "--db-host 127.0.0.1"
    - "--db-port 5432"
    - "--db-user {{env `DB_USER` | def `picard` }}"
    - "--db-pass {{env `DB_PASS` }}"

production:
  deploy:
    server:
      # ...
```

Notice the `{{ ... }}` syntax, this is Go's text templating and is being used here to lookup ENV vars and provide
defaults if they are not defined.

### Commands

#### DB

The `db` deploy command manages migrating and rolling back database changes. It expects your migrations to exist
in `./db/sql/` and to be named `version-x.x.x.sql`. Rollback files should live along side the version files and be named
`rollback-x.x.x.sql`.

You'll need to add a configuration to the Forgefile for the `db` subcommand:

```yaml
production:
  db:
    dbuser: username
    dbpass: {{env `DB_PASS`}}
    dbhost: 127.0.0.1
    dbport: 5432
    sslmode: "require"
```

Here are examples of common operations:

```bash
# Execute all pending migrations
forge db --migrate

# Migrate up to a specific version
forge db --migrate --to 2.0.1

# Rollback the most recent migration
forge db --rollback

# Rollback to a specific version
forge db --rolback --to 0.0.1
```

All available options:

Option|Value|Description
------|-----|-----------
migrate | | Migrate the database forward to the target version. By default, This will run all pending migrations.
rollback | | Rollback to the target version. Be default, this will only rollback the most recent version.
to | x.x.x | Set the target version to migrate to or rollback from.
ssl | string | Set the SSL Mode for connecting to the postgres database.

#### Deploy

The `deploy` subcommand abstracts away interacting with various backend-hosting services into a single `forge deploy` command
that is configured by the Forgefile. The interactions with any particular backend is managed by a "shipper" which manages both
deploying to a hosting service and rolling back changes in the case of an error. Forge tries to revert things back as close to the
original state as possible, but that can't be automated 100% of the time, so be sure to double check your stack after a failed deploy!

An example deploy config could look something like this:

```yaml
production:
  deploy:
    server:             # `server` is a target we want to deploy to. The exact name is arbitrary.
      shipper: k8       # `shipper` specifies the plugin that should be used to accomplish the deploy.
      opts:             # `opts` specifies a list of options to be passed to to the shipper. Its contents are dependent on the shipper.
        arg: val
        arg2: val
    client:             # `client` is another target we want to be deployed.
      shipper: gulp-s3
      opts: {}
    # ...
```

  With this configuration, we could deploy our application by running the following command.

```bash
forge deploy --env production
```

##### Kubernetes Configuration

To configure `forge deploy` to update a kubernetes cluster, you will need a configuration similar to the following:

```yaml
# ...
qa: # <- The environment
  deploy:
    server:          # <- This name is arbitrary and represents the deploy target.
      shipper: k8    # <- Specify the Kubernetes Shipper plugin
      opts:
        server: https://myclusterhost  # <- The host for the Kubernetes cluster.
        ca: PEMENCODEDCA               # <- The certificate authority for the cluster's SSL.
        token: SERVICEACCOUNTTOKEN     # <- The access token for the K8 service account.
```

A complete list of available `opts` for the Kubernetes shipper can be found in the following tables:

| Name        | Required | Value                                                        |
|-------------|----------|--------------------------------------------------------------|
| server      | Yes      | http(s)://hostname(:port)                                    |
| ca          |          | The PEM encoded Certificate Authority for the Kubernetes SSL |
| caFile      |          | The path to the PEM encoded Certificate Authority            |

Options for an authentication scheme must be provided as well. The following tables show the required `opts` for each available authentication scheme. You only need to provide the configuration for one of these.

The easiest and most universal method is to use Kubernetes Service Accounts and access tokens. You can read about Kubernetes Service accounts [here](https://kubernetes.io/docs/admin/service-accounts-admin/), or skip down in this readme to "Generating a Kubernetes Service Account".

| Name        | Required | Value                                            |
|-------------|----------|--------------------------------------------------|
| token       | Yes      | An access token for a Kubernetes service account |

You could also set a Private Key and Certificate:

| Name        | Required | Value                                                                                 |
|-------------|----------|---------------------------------------------------------------------------------------|
| apiKey      | Yes      | The Private Key for Key based Kubernetes authentication, requires `apiCert` to be set |
| apiCert     | Yes      | The Certificate for key based Kubernetes authentication, requires `apiKey` to be set  |

Or, you could load the Private Key and Certificate from their own files (This is how minikube works by default):

| Name        | Required | Value                                                                   |
|-------------|----------|-------------------------------------------------------------------------|
| apiKeyFile  | Yes      | Same as `apiKey` but specifies a path to a file, requires `apiCertFile` |
| apiCertFile | Yes      | Same as `apiCert` but specifies a path to a file, requires `apiKeyFile` |

#### Run

  The run sub-command can be used to pass default arguments to commonly used commands. For example, if you don't want to type all
of the `psql` connection parameters every time you connect to your DB, you could define a `run` target in you Forgefile like this:

```yaml
development:
  run:
    psql:
    - "--host={{env `DB_HOST` | def `localhost` }}"
    - "--dbname={{env `DB_NAME` | def `login_dev` }}"
    - "--username={{env `DB_USER` | def `root` }}"
```

  Then the next time you wanted to connect to your database with psql, you could type:

```bash
forge run psql
```

### Generating a Kubernetes Service Account

Kubernetes allows you to create "Service Accounts" for authenticating bots and machines. To get started, you can use the following commands.

Create the service account object:
```bash
echo "apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
" | kubectl create -f -
```

Add a secret, this generates the tokens:
```bash
echo "apiVersion: v1
kind: Secret
metadata:
  name: my-service-account-secret
  annotations:
    kubernetes.io/service-account.name: my-service-account
    type: kubernetes.io/service-account-token
" | kubectl create -f -
```

Fetch the token:
```bash
kubectl get secrets/my-service-account-secret -o json | jq '.data.token' -r | base64 -D
```

If you do have `jq` installed, this will spit out the token onto the commandline for you. If you do not have `jq`, just run the first part of the command and pull out the "data"."token" attribute. Base64 decode that, and you have your token.
