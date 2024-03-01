import { useState, useContext, useEffect } from 'react'
import { QRCode } from '@/app/components'
import { useRouter } from 'next/router';
import { Grid, Typography } from '@mui/material';
import { getCredentialOffer } from '@/services/issuer';
import { ErrorPopup } from '@/app/components';
import SelectedIssuerContext from '@/contexts/SelectedIssuerContext';

const App = () => {
    const router = useRouter();
    const routerQuery = router.query;

    const [ credentialOffer, setCredentialOffer ] = useState('')
    const [error, setError] = useState<string | null>(null);

    const { selectedIssuerContext } = useContext(SelectedIssuerContext);
    useEffect(() => {
      if (!selectedIssuerContext) {
        router.push('/');
        return;
      }
    }, [selectedIssuerContext, router]);

    useEffect(() => {
        const fetchCredentialOffer = async () => {
            try {
                const offer = await getCredentialOffer(
                  routerQuery.issuer as string, 
                  routerQuery.subject as string,
                  routerQuery.claimId as string)
                setCredentialOffer(offer)
            } catch (error) {
                setError(`Failed to fetch credential offer: ${error}`);
            }
        }

        fetchCredentialOffer()
    }, [routerQuery])
    
    return (
        <Grid container 
              direction="column" 
              justifyContent="flex-start" 
              alignItems="center"
              height="100%">
          {error && <ErrorPopup error={error} />}

          <Grid alignItems="center" item xs={3}>
            <Typography variant="h1">
              Scan QR for fetch credential
            </Typography>
          </Grid>
          <Grid alignItems="center" item xs={3}>
            <QRCode value={JSON.stringify(credentialOffer)}/>
          </Grid>
        </Grid>
    );
}

export default App;