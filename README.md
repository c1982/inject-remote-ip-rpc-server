# inject-remote-ip-rpc-server

This example automatically adds the remote ip address into the incoming json PRC message.

I created a new codec and added a function that injects the ip address into the incoming message.
https://github.com/c1982/inject-remote-ip-rpc-server/blob/main/server/custom_codec.go#L90
