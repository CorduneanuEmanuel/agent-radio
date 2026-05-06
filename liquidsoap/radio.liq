#!/usr/bin/liquidsoap

settings.init.allow_root := true
source = mksafe(playlist(reload=5, reload_mode="watch", "/music/"))

output.icecast(
  %mp3,
  host = "localhost",
  port = 8000,
  password = "mds",
  mount = "/radio",
  source
)
