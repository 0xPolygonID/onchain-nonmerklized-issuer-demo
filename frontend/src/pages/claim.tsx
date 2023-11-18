'use client'

import React, { useState } from 'react';
import Web3 from 'web3';
import { useRouter } from 'next/router';
import { Grid, Box, Typography, Button, Backdrop, CircularProgress} from '@mui/material';
import JSONPretty from 'react-json-pretty';
import contractABI from './abi.json';
import { DID } from '@iden3/js-iden3-core'
import { v4 as uuidv4 } from 'uuid';

const App = () => {
  const router = useRouter();
  const routerQuery = router.query;
  const [account, setAccount] = useState('');
  const [claim, setClaim] = useState({});
  const [isLoaded, setIsLoaded] = useState(false);
  const [onchainCredentialUUID, setOnchainCredentialUUID] = useState(uuidv4());

  const initMetaMask = async () => {
    try {
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      setAccount(accounts[0]);
    } catch (err: any) {
      if (err.code === 4001) {
        console.error('User rejected request');
      } else {
        console.error('Error connecting to MetaMask:', err);
      }
    }
  };

  const getBalance = async () => {
    if (account) {
      const web3 = new Web3(window.ethereum);
      const balanceWei = await web3.eth.getBalance(account);
      const balanceGwei = Math.floor(parseFloat(web3.utils.fromWei(balanceWei, 'gwei')));
      
      const contractAddress = '0x85256776C5B1Bd94C066076caAA3e94Abb20aE56';
      const onchainNonMerklizedIssuer = new web3.eth.Contract(contractABI, contractAddress);

      const userDID = await DID.parse(routerQuery.userID as string)
      const userId = await DID.idFromDID(userDID)

      setIsLoaded(true);
      await onchainNonMerklizedIssuer.methods.issueCredential(
        userId.bigInt(), 
        onchainCredentialUUID,
      ).send({from: account});

      
      const rawClaim = await onchainNonMerklizedIssuer.methods.getUserClaim(
        userId.bigInt(), 
        onchainCredentialUUID,
      ).call({from: account});

      console.log('rawClaim', rawClaim);

      const strClaims = rawClaim.claim.map((claim: any) => claim.toString());
      
      const response = await fetch(`http://localhost:3333/api/v1/identities/did:polygonid:polygon:mumbai:2qCU58EJgrELzhAAX2JWfa5L6nND3fzqhisj84wdiq/claims`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          credentialId: onchainCredentialUUID,
          schemaJsonLd: rawClaim.schemaURL,
          schemaUrlJson: rawClaim.schemaJSON,
          credentialType: rawClaim.credentialType,
          claim: strClaims,
        }),
      });

      const data = await response.json();

      const credentialResponse = await fetch(`http://localhost:3333/api/v1/identities/did:polygonid:polygon:mumbai:2qCU58EJgrELzhAAX2JWfa5L6nND3fzqhisj84wdiq/claims/${data.id}`);
      const credential = await credentialResponse.json();

      console.log('credential', credential);

      router.push(`/offer?claimId=${data.id}&issuer=${credential.issuer}&subject=${credential.credentialSubject.id}`);

      setIsLoaded(false);
    }
  };

  return (
    <Grid container 
      direction="column" 
      justifyContent="center" 
      alignItems="center"
      height="100%"
    >
    {
      !account && (
        <Box textAlign="center">
          <Typography variant="h6">
            Balance claim for user {routerQuery.userID}
          </Typography>
          <Button onClick={initMetaMask} variant="contained" size="large">
            Connect MetaMask
          </Button>
        </Box>
      )
    }  

    {account && claim.credentialSchema === undefined &&  (
      <Grid container direction="column" alignItems="center" textAlign="center">
        <Typography variant="h6">
          Account: {account}
        </Typography>
        <Button onClick={getBalance} variant="contained" size="large">
          Get Balance GWEI
        </Button>
      </Grid>
    )}
  
    {claim.credentialSchema !== undefined && (
      <Grid container direction="column" alignItems="center" textAlign="center">
        <Typography variant="h6" textAlign="left">
            Claim content:
        </Typography>
        <Grid textAlign="left">
          <Box alignItems="left">
            <JSONPretty 
            id="json-pretty" 
            style={{
              fontSize: "1.3em",
            }} 
            data={JSON.stringify(claim)}
            theme={jsonStyle}
          ></JSONPretty>
          </Box>
        </Grid>
        <Button onClick={claimToBalance} variant="contained" size="large">
          Get claim
        </Button>
      </Grid>
    )}
  
    <Backdrop
      sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
      open={isLoaded}
    >
      <CircularProgress color="inherit" />
    </Backdrop>
  </Grid>  
  );
};

const jsonStyle = {
  main: 'line-height:1.3;color:#66d9ef;background:#272822;overflow:auto;',
  error: 'line-height:1.3;color:#66d9ef;background:#272822;overflow:auto;',
  key: 'color:#f92672;',
  string: 'color:#fd971f;',
  value: 'color:#a6e22e;',
  boolean: 'color:#ac81fe;',
}

export default App;
