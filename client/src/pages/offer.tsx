'use client';

import { useState, useContext, useEffect } from 'react'
import { QRCode } from '@/app/components'
import { useRouter } from 'next/router';
import { Grid, Typography } from '@mui/material';
import SelectedIssuerContext from '@/contexts/SelectedIssuerContext';
import { v4 as uuidv4 } from 'uuid';
import { PROTOCOL_CONSTANTS } from '@0xpolygonid/js-sdk'

// TODO(illia-korotia): @0xpolygonid/js-sdk v1.9.x has problem with nextjs.
// These constants should be imported from @0xpolygonid/js-sdk
const mediaTypePlainMessage = "application/iden3comm-plain-json";
const credentialOnchainOfferMessageType = "https://iden3-communication.io/credentials/1.0/onchain-offer";

const App = () => {
  const router = useRouter();
  const routerQuery = router.query;

  // TODO(illia-korotia): use type CredentialsOnchainOfferMessage after fixing @0xpolygonid/js-sdk
  const [offer, setOffer] = useState<any>({});

  const { selectedIssuerContext } = useContext(SelectedIssuerContext);

  useEffect(() => {
    if (!selectedIssuerContext) {
      router.push('/');
      return;
    }

    const addHexPrefix = (s: string): string => {
      if (s.includes('0x') || s.includes('0X')) {
        return s;
      }
      return `0x${s}`;
    };

    setOffer(
      {
        id: uuidv4(),
        typ: mediaTypePlainMessage,
        type: credentialOnchainOfferMessageType,
        thid: uuidv4(),
        body: {
          credentials: [{
            id: routerQuery.claimId as string,
            description: "Non-zero balance credential",
          }],
          transaction_data: {
            contract_address: addHexPrefix(routerQuery.contractAddress as string),
            method_id: "0x37c1d9ff",
            chain_id: 80002,
            network: "polygon-amoy",
          }
        },
        from: routerQuery.issuer as string,
        to: routerQuery.subject as string,
      }
    );
  }, [selectedIssuerContext, routerQuery, router]);

  return (
    <Grid container 
      direction="column" 
      justifyContent="flex-start" 
      alignItems="center"
      height="100%">
      <Grid alignItems="center" item xs={3}>
        <Typography variant="h1">
          Scan QR for fetch credential
        </Typography>
      </Grid>
      <Grid alignItems="center" item xs={3}>
        <QRCode value={JSON.stringify(offer)}/>
      </Grid>
    </Grid>
  );
}

export default App;