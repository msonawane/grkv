
grkv uses grpc and memberlist to create an eventually replicated cluster of badger nodes.

writes ( set and deletes ) are propogated with use of GRPC calls. no guaranties of success
GET call tries to collect data from remote node for any keys current node has no data for.

can be used for local embeded storage / cache. should not be primary data storage.
