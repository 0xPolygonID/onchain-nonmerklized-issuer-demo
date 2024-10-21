import { useState, useEffect, useContext } from 'react';
import { QRCode, ErrorPopup } from '@/app/components';
import { useRouter } from 'next/router';
import { Typography } from '@mui/material';
import Grid from '@mui/material/Unstable_Grid2';
import SelectedIssuerContext from '@/contexts/SelectedIssuerContext';
import { produceAuthQRCode, checkAuthSessionStatus } from '@/services/issuer';

const MyPage = () => {
  const router = useRouter();

  const { selectedIssuerContext } = useContext(SelectedIssuerContext);

  const [qrCodeData, setQrCodeData] = useState({});
  const [sessionID, setSessionID] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!selectedIssuerContext) {
      router.push('/');
      return;
    }

    const fetchAuthQRCode = async () => {
      try {
        const { sessionId, data } = await produceAuthQRCode(selectedIssuerContext);
        setQrCodeData(data);
        setSessionID(sessionId);
      } catch (error) {
        setError(`Failed to fetch QR code: ${error}`);
      }
    };

    fetchAuthQRCode();
  }, [selectedIssuerContext, router]);

  useEffect(() => {
    if (!sessionID) {
      return;
    }

    
    let interval: NodeJS.Timeout;
    const checkStatus = async () => {
      try {
        const response = await checkAuthSessionStatus(sessionID);
        if (response && response.id !== null) {
          clearInterval(interval);
          router.push(`/claim?userID=${response.id}` );
        }
      } catch (error) {
        setError(`Failed to check session status: ${error}`);
      }
    };

    interval = setInterval(checkStatus, 2000);
    return () => clearInterval(interval);
  }, [sessionID, router]);

  return (
    <Grid
      container
      spacing={0}
      direction="column"
      alignItems="center"
      justifyContent="center"
      sx={{ minHeight: '100vh' }}
    >
      {error && <ErrorPopup error={error} />}

      <Grid xs={12} style={{marginTop: '-30px'}}>
        <Typography textAlign="center" variant='h2'>
          Use PrivadoID mobile app to scan this QR code
        </Typography>
      </Grid>
      <Grid alignItems="center" xs={12} style={{marginTop: '30px'}}>
        <QRCode value={JSON.stringify(qrCodeData)}/>
      </Grid>
    </Grid>
  );
};

export default MyPage;