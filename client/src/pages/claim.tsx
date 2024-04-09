'use client'

import React, { useState, useContext, useEffect } from 'react';
import { useRouter } from 'next/router';
import { Grid, Box, Typography, Button, Backdrop, CircularProgress} from '@mui/material';
import { selectMetamaskWallet } from '@/services/metamask';
import { 
  issueCredential, 
  getUserCredentialIds, 
} from '@/services/onchainIssuer';
import { Selecter, ErrorPopup } from '@/app/components';
import SelectedIssuerContext from '@/contexts/SelectedIssuerContext';
import { DID, Id } from '@iden3/js-iden3-core';
import { Hex } from '@iden3/js-crypto';

interface IssuerInfo {
  did: DID;
  id: Id;
  address: string;
};

interface UserInfo {
  did: DID;
  id: Id;
};

const App = () => {
  const router = useRouter();
  const routerQuery = router.query;

  const [userCredentialIdsList, setUserCredentialIdsList] = useState<string[]>([]);
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');
  const [issuerInfo, setIssuerInfo] = useState<IssuerInfo|null>(null);
  const [userInfo, setUserInfo] = useState<UserInfo|null>(null);
  const [error, setError] = useState<string | null>(null);
  const [metamaskWalletAddress, setMetamaskwalletAddress] = useState('');
  const [isLoaded, setIsLoaded] = useState(false);

  

  const { selectedIssuerContext } = useContext(SelectedIssuerContext);
  useEffect(() => {
    if (!selectedIssuerContext) {
      router.push('/');
      return;
    }

    const issuerDid = DID.parse(selectedIssuerContext);
    const issuerId = DID.idFromDID(issuerDid);
    const issuerAddress = Hex.encodeString(Id.ethAddressFromId(issuerId));
    setIssuerInfo({ did: issuerDid, id: issuerId, address: issuerAddress });

    const userDid = DID.parse(router.query.userID as string);
    const userId = DID.idFromDID(userDid);
    setUserInfo({ did: userDid, id: userId });

    getUserCredentialIds(issuerAddress, userId)
      .then((credentials) => {
        setUserCredentialIdsList(credentials.reverse());
      })
      .catch((error) => {
        setError(`Failed to get user credentials: ${error}`);
      })

  }, [selectedIssuerContext, router]);

  const getMetamaskWallet = async () => {
    try {
      const wallet = await selectMetamaskWallet();
      setMetamaskwalletAddress(wallet.address);
    } catch (error) {
      setError(`Failed to get wallet address: ${error}`);
    }
  };

  const selectCredential = async (selecedCredentialId: string) => {
    setSelectedCredentialId(selecedCredentialId);
  }

  const fetchCredential = async () => {
    if (!issuerInfo || !userInfo) {
      setError('Issuer or user info is not defined');
      return;
    }
    const contractAddress = issuerInfo?.address;
    router.push(`/offer?claimId=${selectedCredentialId}&issuer=${selectedIssuerContext}&subject=${routerQuery.userID as string}&contractAddress=${contractAddress}`);
  }

  const issueOnchainCredential = async () => {
    setIsLoaded(true);
    try {
      if (!issuerInfo || !userInfo) {
        setError('Issuer or user info is not defined');
        return;
      }

      await issueCredential(issuerInfo.address, userInfo.id);
      const credentialIds = await getUserCredentialIds(issuerInfo.address, userInfo.id);
      const lastIssuedCredential = credentialIds[credentialIds.length - 1];
  
      router.push(`/offer?claimId=${lastIssuedCredential}&issuer=${selectedIssuerContext}&subject=${routerQuery.userID as string}&contractAddress=${issuerInfo.address}`);
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
          <Box>

            <Box>
              <Typography variant="h6" textAlign="left">
                Select credential to fetch:
              </Typography>
              <Selecter datalist={userCredentialIdsList} callback={selectCredential} label='Select'/>
              <Box marginTop="15px">
                <Button onClick={fetchCredential} variant="contained" size="large">
                  Fetch
                </Button>
              </Box>
            </Box>

            <Box marginTop="35px">
              <Typography variant="h6" textAlign="left">
                Or issue a new balance credential for user:
              </Typography>
              <Typography variant="body1" textAlign="left">
                {routerQuery.userID}
              </Typography>
              <Box marginTop="15px">
                <Button onClick={getMetamaskWallet} variant="contained" size="large">
                  Connect MetaMask
                </Button>
              </Box>
            </Box>

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

export default App;
