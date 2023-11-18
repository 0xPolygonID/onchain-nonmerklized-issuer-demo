import { use, useState, useEffect } from 'react'
import { QRCode } from '@/app/components'
import { useRouter } from 'next/router';
import { v4 as uuidv4 } from 'uuid';
import { Grid, Typography } from '@mui/material';

const App = () => {
    const router = useRouter();
    const routerQuery = router.query;

    const [ credentialOffer, setCredentialOffer ] = useState('')

    useEffect(() => {
        const getOffer = async () => {
            try {
                const offerResponse = await fetch(`http://localhost:3333/api/v1/identities/${routerQuery.issuer}/claims/offer?subject=${routerQuery.subject}&claimId=${routerQuery.claimId}`)
                setCredentialOffer(JSON.stringify(await offerResponse.json()))
            } catch (e) {
                console.log('failed get offer message ->', e)
            }
        }
        getOffer()
    }, [])
    
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
            <QRCode value={credentialOffer}/>
          </Grid>
        </Grid>
    );
}

export default App;