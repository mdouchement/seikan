<!-- TOC -->

- [1. Overview](#1-overview)
- [2. Session](#2-session)
    - [2.1. Workflow](#21-workflow)
- [3. Noise Protocol handshake](#3-noise-protocol-handshake)
    - [3.1. Specifications](#31-specifications)
    - [3.2. Advertising](#32-advertising)
    - [3.3. Chunked stream](#33-chunked-stream)
- [4. Control PDU](#4-control-pdu)
    - [4.1. Payloads](#41-payloads)
        - [4.1.1. error](#411-error)
        - [4.1.2. inbounds](#412-inbounds)
        - [4.1.3. bind_cs](#413-bind_cs)
        - [4.1.4. bind_sc](#414-bind_sc)
- [5. Stream](#5-stream)

<!-- /TOC -->

# 1. Overview

Layers used when a tunnel is opened (after handshake).

| Layer           | Type                 | Description          |
|-----------------|----------------------|----------------------|
| 3. Multiplexing | Yamux                |                      |
| 2. Compression  | Zstandard            | Improve layers 0 & 1 |
| 1. Encryption   | Noise chunked stream |                      |
| 0. Transport    | TCP                  | tcp://               |


# 2. Session

A session is a tunnel for one destination.

The TTL of a session is the TLL of the TCP connection.
If the encryption mechanism fails at any points the server drains and closes the connection.

## 2.1. Workflow

1. TCP connection (always opened by the client)
2. Noise Protocol handshake
3. Control exchanges
4. Streaming

# 3. Noise Protocol handshake

## 3.1. Specifications

Seikan uses Noise `IK` pattern handshake.

Noise Protocol configuration:
- Curve25519 ECDH
- ChaCha20-Poly1305 AEAD
- BLAKE2b hash.

A **failed handshake** results in a closed connection.
Any data that fails AEAD authentication results in a closed connection.

## 3.2. Advertising

The client advertizes the server with its derived identifier then the server knows which public key to use to initiate the `IK` Noise pattern.

To do so, HKDF with blake2b is used:

| Name    | size     | Description               |
|---------|----------|---------------------------|
| salt    | 16 bytes | Key to initialize Blake2b |
| derived | 32 bytes | HKDF output               |


*Security not really matters here since it's the client's identifier, IK handshake is secure.*
It's to always have a 48 bytes length payload with some randomness in the value to make data more difficult to understand for attacker.


## 3.3. Chunked stream

All the encrypted data is sent in chunks through TCP connection.

| Name    | Raw Type          | Type   | Description         |
|---------|-------------------|--------|---------------------|
| size    | 2 bytes BigEndian | uint16 | Payload size        |
| payload | bytes             | []byte | Noise encypted data |


# 4. Control PDU


|    Name    |        Raw Type        |  Type  |     Description    |
|:----------:|:----------------------:|:------:|:------------------:|
| size       | 2 bytes BigEndian      | uint16 | Total PDU size     |
| version    | 1 byte                 | uint8  | PDU version        |
| cid        | 2 bytes BigEndian      | uint8  | Control identifier |
| pid        | null-terminated string | string | PDU identifier     |
| payload    | CBOR                   | struct | Control data       |

The header is formed by `size` + `version` + `cid` + `pid`.

`pid` must be the same for both request and response.


## 4.1. Payloads

### 4.1.1. error

**Only** for response.

control-id: `0x01`

|  Field  |  Type  |        Description       |
|:-------:|:------:|:------------------------:|
| status  | int    | The error status         |
| message | string | The message of the error |

`status` value is based on the HTTP status codes.

### 4.1.2. inbounds

Asked by the client to the server to get the `server -> client` tunnels.


1. Request

control-id: `0x02`

|    Field   |  Type  | Description |
|:----------:|:------:|:-----------:|
| identifier | string | Client ID   |

2. Response

control-id: `0x03`

|     Field    |   Type   |      Description      |
|:------------:|:--------:|:---------------------:|
| inbounds     | []string | Inbounds addresses    |

### 4.1.3. bind_cs

Used by the client to open a tunnel.
This tunnel will accept a bidirectional connection from client to server (the client start the listener).
**After this control the stream begins.**

1. Request

control-id: `0x04`

|    Field    |  Type  |     Description     |
|:-----------:|:------:|:-------------------:|
| identifier  | string | Client ID           |
| address     | string | Server side address |

2. Response

control-id: `0x05`

|     Field    |   Type   |        Description       |
|:------------:|:--------:|:------------------------:|
|              |          |                          |

### 4.1.4. bind_sc

Used by the client to open a tunnel.
This tunnel will accept a bidirectional connection from server to client  (the server start the listener).
**After this control the stream begins.**

1. Request

control-id: `0x06`

|    Field    |  Type  |     Description     |
|:-----------:|:------:|:-------------------:|
| identifier  | string | Client ID           |
| address     | string | client side address |

2. Response

control-id: `0x07`

|     Field    |   Type   |        Description       |
|:------------:|:--------:|:------------------------:|
|              |          |                          |


# 5. Stream

It uses [Yamux](https://github.com/hashicorp/yamux) following this [spec](https://github.com/hashicorp/yamux/blob/master/spec.md).
Its used for multiplexing requests and easily implementing the `bind_sc` feature.