'use client'

import React, { useState, useContext, useEffect } from 'react';
import { useRouter } from 'next/router';
import { Grid, Box, Typography, Button, Backdrop, CircularProgress} from '@mui/material';
import { selectMetamaskWallet } from '@/services/metamask';
import { 
  issueCredential, 
  userCredentialIds, 
  getCredential, 
  getOnchainIssuerVersion, 
  isOnchainIssuerInterfaceImplemented,
} from '@/services/onchainIssuer';
import { convertClaim } from '@/services/issuer';
import { ErrorPopup } from '@/app/components';
import SelectedIssuerContext from '@/contexts/SelectedIssuerContext';
import { DID, Id } from '@iden3/js-iden3-core';
import { Hex } from '@iden3/js-crypto';

const App = () => {
  const router = useRouter();
  const routerQuery = router.query;
  

  const { selectedIssuerContext } = useContext(SelectedIssuerContext);
  useEffect(() => {
    if (!selectedIssuerContext) {
      router.push('/');
      return;
    }
  }, [selectedIssuerContext, router]);

  const [error, setError] = useState<string | null>(null);

  const [metamaskWalletAddress, setMetamaskwalletAddress] = useState('');
  const getMetamaskWallet = async () => {
    try {
      const wallet = await selectMetamaskWallet();
      setMetamaskwalletAddress(wallet.address);
    } catch (error) {
        setError(`Failed to get wallet address: ${error}`);
    }
  };

  const [isLoaded, setIsLoaded] = useState(false);
  const issueOnchainCredential = async () => {
    setIsLoaded(true);
    try {
      // extract contract address for issuer did
      const issuerDid = DID.parse(selectedIssuerContext);
      const issuerId = DID.idFromDID(issuerDid);
      const contractAddress = Hex.encodeString(Id.ethAddressFromId(issuerId));

      // extract user id from user did
      const userDid = routerQuery.userID as string;
      const userId = DID.idFromDID(DID.parse(userDid));

      const isImplemented = await isOnchainIssuerInterfaceImplemented(contractAddress);
      if (!isImplemented) {
        throw new Error('Onchain issuer interface not implemented');
      }

      await issueCredential(contractAddress, userId);
      const credentialIds = await userCredentialIds(contractAddress, userId);
      const onchainCredential = await getCredential(contractAddress, userId, credentialIds[credentialIds.length - 1]);
      const version = await getOnchainIssuerVersion(contractAddress);
      const verifiableCredentialId = await convertClaim(selectedIssuerContext, onchainCredential, version);
  
      router.push(`/offer?claimId=${verifiableCredentialId}&issuer=${selectedIssuerContext}&subject=${routerQuery.userID as string}`);
    } catch (error) {
      setError(`Failed to issue onchain credential: ${error}`);
    } finally {
      setIsLoaded(false);
    }
  }
  

  return (
    <Grid container 
      direction="column" 
      justifyContent="center" 
      alignItems="center"
      height="100%"
    >
    {error && <ErrorPopup error={error} />}
    
    {
      !metamaskWalletAddress && (
        <Box textAlign="center">
          <Typography variant="h6">
            Balance claim for user {routerQuery.userID}
          </Typography>
          <Button onClick={getMetamaskWallet} variant="contained" size="large">
            Connect MetaMask
          </Button>
        </Box>
      )
    }  
 
    {metamaskWalletAddress && (
      <Grid container direction="column" alignItems="center" textAlign="center">
        <Typography variant="h6">
          Wallet: {metamaskWalletAddress}
        </Typography>
        <Button onClick={issueOnchainCredential} variant="contained" size="large">
          Issue onchain credential
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
