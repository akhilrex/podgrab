#!/bin/sh
addgroup -S -g $PGID user
adduser \
    --disabled-password \
    --gecos "" \
    --ingroup "user" \
    --no-create-home \
    --uid "$PUID" \
    "user"
echo Running as PID $PUID and GID $PGID.
echo Starting Podgrab...
su -s /bin/sh -c "./app" user