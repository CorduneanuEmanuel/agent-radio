#!/usr/bin/liquidsoap
settings.init.allow_root := true

def on_new_track(m) =
  ignore(http.post("http://167.172.171.185:8080/track-started",
    headers=[("Content-Type", "application/json")],
    data=json.stringify(m)))
end

s = mksafe(playlist(reload=5, reload_mode="watch", "/music/"))
s.on_track(on_new_track)

output.icecast(
  %mp3,
  host = "localhost",
  port = 8000,
  password = "mds",
  mount = "/radio",
  s
)
