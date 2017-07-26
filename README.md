# mm-client
A front end to the mm structure.

## config
By default mm-client reads ~/.config/mm/rooms.json & accounts.json.
This directory is configurable with the -c option.

Note escape codes `\033` must be escaped in json.

rooms.json
``` json
{
	"!RmLAUQhJdhvxTpzIOm:matrix.org": "\\033[1;36msarah\\033[0m",
	"!inLIeOyAtLMFVYBttb:server.org": "family",
}
```

accounts.json
``` json
{
        "@sarah:matrix.org":       "sarah",
	"@john:server.org":        "john",
}
```
