# Seikan

Seikan is client/server application that enables to create bidirectional TCP tunnels leveraging Noise Protocol.

It uses the excellent yamux package to multiplex connections between server and client.

## Usage

- Create a new identities (client & server)

```sh
$ seikan identity
secret: sk-M1KTqaRwiJGDVf9vzP6yJoaArJ3DW7dCCq1qPXRxdiS
public: pk-GEdcuHcNyapH3K52JuURzaUXFYrTDk1tQj4EhZa9WDqX
```

- Setup both `client.yml` & `server.yml`
- Run `seikan server -c server.yml`
- Run `seikan client -c client.yml`

### Features

- Client to server bidirectional TCP tunnel
- Server to client bidirectional TCP tunnel
- Encrypted using the Noise Protocol


### Technologies / Frameworks

- [Cobra](https://github.com/spf13/cobra)
- [Noise](https://github.com/flynn/noise)
- [Yamux](https://github.com/hashicorp/yamux)
- [Compress](https://github.com/klauspost/compress) for zstandard compression
- [CBOR](https://github.com/fxamacker/cbor)


## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
