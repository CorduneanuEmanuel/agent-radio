#!/usr/bin/liquidsoap
settings.init.allow_root := true

def on_new_track(m) =
  ignore(http.post("http://localhost:8080/track-started",
    headers=[("Content-Type", "application/json")],
    data=json.stringify(m)))
end

source = mksafe(playlist(reload=5, reload_mode="watch", "/music/"))
source = source.on_track(on_new_track)

output.icecast(
  %mp3,
  host = "localhost",
  port = 8000,
  password = "mds",
  mount = "/radio",
  source
)
