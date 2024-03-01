import { FunctionComponent } from "react";
import QRCode from 'qrcode.react';
import { Button, Container, Box } from '@mui/material';

interface QRCodeData {
    value: string;
}

const Code: FunctionComponent<QRCodeData> = (props) => {
  const { value } = props;
  return (
    <Container style={{width: '350px'}}>
      <Box>
        <QRCode level={"L"} size={350} value={value} />
        <Button onClick={() => dispachEvent(value)} 
            variant="contained" size="large" style={{width: '350px', marginTop: '15px'}}>Polygon ID</Button>
      </Box>
    </Container>
  );
};

const dispachEvent = async (value: string) => {
  const msg = btoa(value);
  const hrefValue = `iden3comm://?i_m=${msg}`;

  const _authEvent = new CustomEvent('authEvent', { detail: hrefValue });
  document.dispatchEvent(_authEvent);
}

export default Code;
