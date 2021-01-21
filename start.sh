#!/bin/sh
chown -R $PUID:$PGID /config /assets /api
echo Running as PID $PUID and GID $PGID.
echo Starting Podgrab...
su-exec $PUID:$PGID ./app