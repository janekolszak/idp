# Hydra IdP integration example
This is an example integration of a simple Identity Provide and [Hydra](https://github.com/ory-am/hydra)
The IdP doesn't authenticate, it grants access to every request. Normally you'd add a step for user authentication, like checking user:password credentials.

## Instructions
Example uses **docker-compose** for orchestration, it starts two containers:
- **hydra** with in-memory database, listening on https://localhost:4444
- **idp** listening on http://localhost:3000 (this should be https in production)

#### Console 1:
Start Hydra and browse it's logs. Copy the client's credentials, you'll need them in Console 3.
``` bash
docker-compose up hydra
```

#### Console 2:
Start IdP and browse it's logs
``` bash
docker-compose up idp
```

#### Console 3
Perform some experiments like:
``` bash
# Pass the credentials from Console 1
hydra connect

# Might not work with Chrome, but works with Firefox
hydra token user --skip-tls-verify
```
