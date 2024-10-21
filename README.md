# Onchain non-merklized issuer demo

> **NOTE**: This is a demo only. Do not use it in a production environment.

### Description

This is a demo frontend page that communicates with the [non-merklized on-chain issuer](https://github.com/0xPolygonID/contracts/blob/main/contracts/examples/BalanceCredentialIssuer.sol) to issuer non-merklized credential. The **on-chain non-merklized issuer** can use information from a blockchain (such as balance, token ownership, etc.) to issue a credential directly on the blockchain. This approach is decentralized and trustless - no need to trust an issuer to act honestly, because it's enforced by the smart contract and auditable on chain. But it comes with a few limitations: max 4 data fields in the credential and data is public. More about [non-merklized credentials](https://docs.iden3.io/protocol/non-merklized/). You can also consider the [merklized on-chain issuer](https://github.com/0xPolygonID/onchain-merklized-issuer-demo) solution if you need a centralised issuer without a limit on the number of fields.

### Quick Start Installation 

**Requirements:**
- Docker
- Docker-compose
- Ngrok

**Steps to run:**

1. Deploy the [non-merklized on-chain issuer](https://github.com/0xPolygonID/contracts/blob/main/contracts/examples/BalanceCredentialIssuer.sol). [Script to deploy](https://github.com/0xPolygonID/contracts/blob/main/scripts/deployBalanceCredentialIssuer.ts) or use the [npm command](https://github.com/0xPolygonID/contracts/blob/d308e1f586ea177005b34872992d16c3cb20e474/package.json#L62). 

2. Copy `.env.example` to `.env`
    ```sh
    cp .env.example .env
    ```

3. Run `ngrok` on 8080 port.
    ```sh
    ngrok http 8080
    ```

4. Use the utility to calculate the issuerDID from the smart contract address:
    ```bash
    go run utils/convertor.go --contract_address=<ADDRESS_OF_ONCHAIN_ISSUER_CONTRACT>
    ```
    Available flags:
    - `contract_address` - contract address that will convert to did
    - `network` - network of the contract. Default: **polygon**
    - `chain` - chain of the contract. Default: **amoy**

5. Fill the `.env` config file with the proper variables:
    ```bash
    SUPPORTED_STATE_CONTRACTS="80002=<AMOY_STATE_CONTRACT_ADDRESS>,21000=<PRIVADO_STATE_CONTRACT_ADDRESS>"
    SUPPORTED_RPC="80002=<RPC_POLYGON_AMOY>,21000=<RPC_PRIVADO_MAIN>"
    ISSUERS="<ISSUER_DID>"
    EXTERNAL_HOST="<NGROK_URL>"
    ```
    `ISSUERS` supports an array of issuers in the format `"issuerDID1,issuerDID2"`

6. Use the docker-compose file:
    ```bash
    docker-compose build
    docker-compose up -d
    ```

7. Open: http://localhost:3000

## How to verify the non zero balance claim:
1. Visit [https://tools.privado.id/query-builder](https://tools.privado.id/query-builder).
2. Build the next verification request:
    ```text
    URL to JSON-LD Context: https://gist.githubusercontent.com/ilya-korotya/660496c859f8d31a7d2a92ca5e970967/raw/6b5fc14fe630c17bfa52e05e08fdc8394c5ea0ce/non-merklized-non-zero-balance.jsonld
    
    Schema type: Balance

    Attribute field: balance | address

    Proof type: Merkle Tree Proof (MTP)

    Circuit ID: any

    Query type: Selective disclosure
    ```
3. Press: Create query
4. Scan the QR code with PrivadoID application

## License

onchain-nonmerklized-issuer-demo is part of the 0xPolygonID project copyright 2024 ZKID Labs AG

This project is licensed under either of

- [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) ([`LICENSE-APACHE`](LICENSE-APACHE))
- [MIT license](https://opensource.org/licenses/MIT) ([`LICENSE-MIT`](LICENSE-MIT))

at your option.
