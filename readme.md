# FirstBot

Inspired by the bot present in `steve_boots` twitch channel. People can race to say !first when the stream goes live and it tracks who got it along with some history stats. Just a funw ay to encourage some chat at the top of a stream.

This is definitely not "production-ready" code but was a chance for me to learn about the EventSub websocket interface. As I haven't seen any golang implementations for it, figured I'd put this as open source for others to see. In terms of not being production ready, I do not properly handle revokation or expiration of tokens very well, though it attempts a refresh of the access token every time it starts up. So it kinda sucks but letting it die when something is revoked/fails and bringing it back up is the way to go right now. It does however handle reconnect flows.


## For the Developers

If you're just interested in the usage of the EventSub API, its `pkg/eventsub` that is of interest to you. Also `pkg/eventsub/subscriptions` its a bit of a framework you can use. Implement the subcription interface to get callbacks and then you can register your subscription structure and get callbacks. I've only implemented `stream.online` and `stream.offline` here. 

It does require a Twitch Application Client ID and Secret along with access and refersh tokens for a specific user who has authorized that app. This is because EventSub requires using a user-level access token when using the websocket.

For testing if you want to use the twitch cli to host a fake EventSub service, you can use the environment variables `EVENTSUB_SOCKET_URL` and `EVENTSUB_SUBSCRIPTION_URL` to override the live URLs.

## Setup/Configuration

There are three arguments/files that are required.

**`-config=<filename>`** - The core configuration file. Documented in [cmd/firstbot/config.go](cmd/firstbot/config.go)

```json
{
  "broadcaster":"62635947",
  "chat": {
    "username":"earlybirdbot",
    "token":"",
    "channel":"sodawavelive"
  }
}
```

**`-twitch=<filename>`** - Is the twitch authorization tokens and such. Documented in [pkg/twitchclient/types.go](pkg/twitchclient/types.go)

You can get a client id and client secret by creating an application on [dev.twitch.tv](https://dev.twitch.tv/).

You then need to follow the normal oauth flow to get an access and refresh token. This is documented [here](https://dev.twitch.tv/docs/authentication/) or you can try to use [TwitchTokenGenerator](https://twitchtokengenerator.com/) if you don't want to setup a catcher for that. This application assumes you can get a token first, but can refresh them and will update the file when it does.

```json
{
  "application": {
    "client_id":"<client id>",
    "client_secret":"<client secret>>"
  },
  "access_tokens": {
    "access_token":"<access token>>",
    "refresh_token":"<refresh token>"
  }
}
```

**`-state=<filename>`** - This file doesn't need to exist already, it will be created. If no argument is provided it'll be written to state.json. This is just where it saves information like how many times someone has claimed !first.

