# Gobo Stick
This will start a local server for handling "Sign in with Google" requests at /api/sign-in-with-google.

## .env
You **must** set following env variables. (See .env.sample)
```
APP_URL=
SESSION_NAME=

HTTP_PORT=

CLIENT_ID=
CLIENT_SECRET=
REDIRECT_URL=

DB_DRIVER=
DB_HOST=
DB_PORT=
DB_NAME=
DB_USER=
DB_PASSWORD=
REDIS_HOST=
REDIS_PORT=
```

## OAuth2
Follow these [instructions](https://developers.google.com/identity/protocols/OAuth2) to create your OAuth client credentials.
Once you obtain your credentials, set `CLIENT_ID`, `CLIENT_SECRET`, `REDIRECT_URL`.

## Database
Must be running mysql with a *user* table.

## Redis
Must be running redis.

## Getting Started
```
docker-compose up --build -d
```
