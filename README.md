# Onchain Non-Merkleized Issuer Demo
This demo illustrates how to use a non-merkleized onchain issuer.

## Frontend
The frontend client communicates with the auth server for user authentication.

**How to run:**
1. Execute the following commands:
    ```bash
    cd frontend
    npm install
    npm run dev
    ```
To change the onchain issuer contract, calculate the issuer DID from the contract address using this Go code snippet.

## Auth-Server
The Auth-Server is used for user authorization and retrieving user's DID.
**How to Run:**
1. Populate the .env file with your onchain non-merkleized contract information:
    ```bash
    export ONCHAIN_ISSUER_CONTRACT_ADDRESS=0x85256776C5B1Bd94C066076caAA3e94Abb20aE56 # onchain-issuer contract address
    export ONCHAIN_ISSUER_CONTRACT_BLOCKCHAIN=polygon # blockchain name
    export ONCHAIN_ISSUER_CONTRACT_NETWORK=mumbai # network name
    ```
2. `source .env`
3. Create a resolvers.settings.yaml file with the following content:
    ```yaml
    polygon:mumbai: # network prefix
      contractState: 0x134B1BE34911E39A8397ec6289782989729807a4 # state contract address in mumbai
      networkURL: <node_rpc_url>
    ```
4. Run the auth-server:
    ```bash
    cd auth-server
    go run . --dev
    ```
## Backend Server
The Backend Server converts [core claim](https://docs.iden3.io/protocol/claims-structure/) from the onchain issuer to a verifiable credential.
**How to Run:**
1. Populate the .env file:
    ```bash
    export NODE_RPC_URL=<node_rpc_url>
    ```
2. `source .env`
3. Populate the resolvers.settings.yaml file:
    ```yaml
    polygon:mumbai: # network prefix
      contractState: 0x134B1BE34911E39A8397ec6289782989729807a4 # state contract address
      networkURL: <infura_url>
    ```
4. Run the backend:
    ```bash
    cd backend
    docker-compose up -d onchainmongo
    go run . --dev
    ```
