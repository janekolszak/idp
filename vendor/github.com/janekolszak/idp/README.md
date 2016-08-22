# Identity Provider (IdP) for Hydra
[![Build Status](https://travis-ci.org/janekolszak/idp.svg?branch=master)](https://travis-ci.org/janekolszak/idp)
[![Coverage Status](https://coveralls.io/repos/github/janekolszak/idp/badge.svg?branch=master)](https://coveralls.io/github/janekolszak/idp?branch=master)

**Under development**

If you're looking for an example IdP integration with Hydra - it's [here](https://github.com/janekolszak/hydra-idp-go).

Let me know if you wan't to take part in the development to speed things up. Join the conversation on Gitter: [![Gitter](https://img.shields.io/gitter/room/nwjs/nw.js.svg?maxAge=2592000)](https://gitter.im/janekolszak/idp)

## About
Writing a general, all purpose Identity Provider is beyond me.
Instead I want to provide this little playground with different tools that you can use to create your own ideal IdP.

## Running the example:
#### Console 1:
Start Hydra and browse it's logs. Copy the client's credentials, you'll need them in Console 3.
``` bash
cd example
docker-compose up hydra
```

#### Console 2:
Start IdP and browse it's logs
``` bash
cd example
docker-compose up idp
```

#### Console 3
Perform some experiments like:
``` bash
# Pass the credentials from Console 1
hydra connect
hydra token user --skip-tls-verify --no-open
# Paste the link to Firefox
```

## TODO:
- Rethinkdb storages
- Login/Logout endpoint
- Register user endpoint
- Encrypting cookies
- Use hydra's client library
- Handle expirtion of remember me cookies
- Handle errors from hydra
- Parsing configuration file in examples or env variables
- Trusted clients that won't trigger asking user to agree upon scopes
- Digest Auth Provider
- Providers should return user id, not username
- Request removing bad cookies in responses
- Verify email
- Reset password
- Use worker pool in sending emails etc.
- Email templates passed via files